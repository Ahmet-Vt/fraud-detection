import { useEffect, useMemo, useState } from 'react'
import { CartesianGrid, Bar, Line, ComposedChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { useScreen } from '../config/ScreenContext'
import type { UserRow } from './UsersTable'

type UserTransactionsPanelProps = {
  user: UserRow
  onBack: () => void
  transactionsBaseUrl?: string
}

type UserTransaction = {
  id: string
  amount: number
  location: string
  transactionStatus: string
  fraudStatus: boolean
  createdAt: string
}

type DailySummaryPoint = {
  bucketStart: string
  requestCount: number
  anomalyCount: number
  amountSum: number
  amountAvg: number
}

const pageSize = 100

function formatDate(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  const dd = String(date.getDate()).padStart(2, '0')
  const mm = String(date.getMonth() + 1).padStart(2, '0')
  const yyyy = String(date.getFullYear())
  return `${dd}:${mm}:${yyyy}`
}

function formatTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--:--:--'
  const hh = String(date.getHours()).padStart(2, '0')
  const mm = String(date.getMinutes()).padStart(2, '0')
  const ss = String(date.getSeconds()).padStart(2, '0')
  return `${hh}:${mm}:${ss}`
}

export default function UserTransactionsPanel({
  user,
  onBack,
  transactionsBaseUrl = '/transactions/user',
}: UserTransactionsPanelProps) {
  const { isXs, isSm, isMd, ui } = useScreen()
  const [transactions, setTransactions] = useState<UserTransaction[]>([])
  const [dailySummary, setDailySummary] = useState<DailySummaryPoint[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)

  useEffect(() => {
    let cancelled = false
    const controller = new AbortController()

    const loadTransactions = async () => {
      setIsLoading(true)
      setError(null)
      try {
        const [transactionsResponse, summaryResponse] = await Promise.all([
          fetch(`${transactionsBaseUrl}/${user.id}`, { signal: controller.signal }),
          fetch(`${transactionsBaseUrl}/${user.id}/summary?bucket=day`, { signal: controller.signal }),
        ])

        if (!transactionsResponse.ok) {
          throw new Error(`transactions request failed with ${transactionsResponse.status}`)
        }
        if (!summaryResponse.ok) {
          throw new Error(`summary request failed with ${summaryResponse.status}`)
        }

        const payload = (await transactionsResponse.json()) as Array<{
          id: string
          amount: number
          location: string
          transaction_status: string
          fraud_status: boolean
          created_at?: string
        }>
        const summaryPayload = (await summaryResponse.json()) as Array<{
          bucket_start: string
          request_count: number
          anomaly_count: number
          amount_sum: number
          amount_avg: number
        }>
        if (cancelled) return
        const mapped = payload
          .map((tx) => ({
            id: tx.id,
            amount: tx.amount,
            location: tx.location,
            transactionStatus: tx.transaction_status,
            fraudStatus: tx.fraud_status,
            createdAt: tx.created_at ?? '',
          }))
          .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())

        const mappedSummary = summaryPayload
          .map((item) => ({
            bucketStart: item.bucket_start,
            requestCount: item.request_count,
            anomalyCount: item.anomaly_count,
            amountSum: item.amount_sum,
            amountAvg: item.amount_avg,
          }))
          .sort((a, b) => new Date(a.bucketStart).getTime() - new Date(b.bucketStart).getTime())

        setTransactions(mapped)
        setDailySummary(mappedSummary)
        setPage(1)
      } catch (err) {
        if (cancelled) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setError('Could not load user transactions')
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }

    void loadTransactions()
    return () => {
      cancelled = true
      controller.abort()
    }
  }, [transactionsBaseUrl, user.id])

  const anomalyCount = useMemo(
    () => transactions.reduce((acc, tx) => acc + (tx.fraudStatus ? 1 : 0), 0),
    [transactions],
  )
  const anomalyRatio = transactions.length === 0 ? 0 : (anomalyCount / transactions.length) * 100

  const chartData = useMemo(
    () =>
      dailySummary.map((point, index) => ({
          pointIndex: index,
          amount: point.amountSum,
          anomalyCount: point.anomalyCount,
          timeLabel: formatDate(point.bucketStart),
          createdAt: point.bucketStart,
        })),
    [dailySummary],
  )

  const totalPages = Math.max(1, Math.ceil(transactions.length / pageSize))
  const pagedTransactions = useMemo(() => {
    const start = (page - 1) * pageSize
    return transactions.slice(start, start + pageSize)
  }, [transactions, page])

  return (
    <div className={`space-y-3 ${ui.stackGap}`}>
      <div className="card bg-base-100 shadow-md">
        <div className={`card-body ${ui.cardBodyPadding}`}>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div className="flex items-center gap-2">
              <button type="button" className={`btn btn-outline ${isMd ? 'btn-xs' : 'btn-sm'}`} onClick={onBack}>
                Back to Users
              </button>
              <h2 className={`card-title ${ui.cardTitleText}`}>
                {user.name} {user.surname}
              </h2>
            </div>
            <span className={`badge badge-error badge-outline ${ui.badgeSize}`}>
              Anomaly Rate: {anomalyRatio.toFixed(2)}%
            </span>
          </div>
          {error && <div className={`alert alert-error ${ui.tableCellText}`}>{error}</div>}
          {isLoading && <div className={`alert alert-info ${ui.tableCellText}`}>Loading user transactions...</div>}
          <div className={`${ui.chartHeight} w-full`}>
            <ResponsiveContainer width="100%" height="100%">
              <ComposedChart
                data={chartData}
                margin={{
                  top: 8,
                  right: isSm ? 8 : 12,
                  bottom: isSm ? 0 : 8,
                  left: isXs ? 12 : isSm ? 16 : 20,
                }}
              >
                {ui.chartGridVisible && <CartesianGrid strokeDasharray="3 3" />}
                <XAxis
                  dataKey="pointIndex"
                  type="number"
                  tickFormatter={(value: number) => chartData[value]?.timeLabel ?? ''}
                  minTickGap={isSm ? 14 : isMd ? 20 : 28}
                />
                <YAxis yAxisId="amount" width={isXs ? 56 : isSm ? 62 : isMd ? 68 : 76} />
                <YAxis
                  yAxisId="anomaly"
                  orientation="right"
                  width={isXs ? 34 : isSm ? 38 : isMd ? 42 : 48}
                  allowDecimals={false}
                />
                <Tooltip
                  labelFormatter={(value) => {
                    const point = chartData[Number(value)]
                    if (!point) return 'Time: --:--:--'
                    return `Time: ${point.timeLabel} | Date: ${formatDate(point.createdAt)}`
                  }}
                />
                <Bar yAxisId="amount" dataKey="amount" fill="#60a5fa" name="Daily Amount Sum" />
                <Line yAxisId="anomaly" type="monotone" dataKey="anomalyCount" stroke="#ef4444" strokeWidth={2} />
              </ComposedChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      <div className="card bg-base-100 shadow-md">
        <div className={`card-body ${ui.cardBodyPadding}`}>
          <h3 className={`card-title ${ui.cardTitleText}`}>Transactions</h3>
          <div className={`${isMd ? 'max-h-[22rem]' : 'max-h-[28rem]'} overflow-auto`}>
            <table className={`table table-zebra ${ui.tableSize}`}>
              <thead>
                <tr className={`sticky top-0 z-10 bg-base-100 ${ui.tableCellText}`}>
                  <th>ID</th>
                  <th>Amount</th>
                  {!isMd && <th>Location</th>}
                  <th>Status</th>
                  <th>Anomaly</th>
                  <th>Date</th>
                </tr>
              </thead>
              <tbody>
                {!isLoading && pagedTransactions.length === 0 ? (
                  <tr>
                    <td colSpan={isMd ? 5 : 6} className="text-center opacity-70">
                      No transactions for this user.
                    </td>
                  </tr>
                ) : (
                  pagedTransactions.map((tx) => (
                    <tr key={tx.id} className={ui.tableCellText}>
                      <td className="font-medium">{tx.id}</td>
                      <td>${tx.amount.toLocaleString()}</td>
                      {!isMd && <td>{tx.location}</td>}
                      <td>{tx.transactionStatus}</td>
                      <td>{tx.fraudStatus ? 'anomaly' : 'approved'}</td>
                      <td>{formatDate(tx.createdAt)}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
          {transactions.length > 0 && (
            <div className="mt-3 flex items-center justify-between">
              <span className={ui.tableCellText}>
                Page {page} / {totalPages} - 100 rows per page
              </span>
              <div className="join">
                <button
                  type="button"
                  className={`btn join-item ${isMd ? 'btn-xs' : 'btn-sm'}`}
                  disabled={page === 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                >
                  Prev
                </button>
                <button
                  type="button"
                  className={`btn join-item ${isMd ? 'btn-xs' : 'btn-sm'}`}
                  disabled={page >= totalPages}
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
