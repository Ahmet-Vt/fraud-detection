import { useEffect, useMemo, useState } from 'react'
import { CartesianGrid, Bar, Line, ComposedChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { apiUserTransactionListUrl, apiUserTransactionSummaryUrl } from '../api/endpoints'
import { useScreen } from '../config/ScreenContext'
import {
  formatCalendarDayHeading,
  formatDateTime,
  localCalendarDayKey,
} from '../lib/formatDateTime'
import type { DateRangePreset } from '../lib/userDetailDateRange'
import { userDetailDateRangeQuery } from '../lib/userDetailDateRange'
import type { UserRow } from './UsersTable'

const DATE_PRESET_OPTIONS: { value: DateRangePreset; label: string }[] = [
  { value: 'today', label: 'Today' },
  { value: 'weekly', label: 'Last 7 days' },
  { value: 'monthly', label: 'This month' },
  { value: 'yearly', label: 'This year' },
  { value: 'all', label: 'All' },
]

type UserTransactionsPanelProps = {
  user: UserRow
}

type UserTransaction = {
  id: string
  amount: number
  location: string
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

type TransactionTableSegment =
  | { kind: 'day'; dayKey: string; label: string }
  | { kind: 'row'; tx: UserTransaction }

const pageSize = 100

function transactionRowsWithDayBreaks(rows: UserTransaction[]): TransactionTableSegment[] {
  const out: TransactionTableSegment[] = []
  let prevDay: string | null = null
  for (const tx of rows) {
    const dayKey = localCalendarDayKey(tx.createdAt)
    if (dayKey !== null && dayKey !== prevDay) {
      prevDay = dayKey
      out.push({ kind: 'day', dayKey, label: formatCalendarDayHeading(tx.createdAt) })
    }
    out.push({ kind: 'row', tx })
  }
  return out
}

export default function UserTransactionsPanel({ user }: UserTransactionsPanelProps) {
  const { isXs, isSm, isMd, ui } = useScreen()
  const [transactions, setTransactions] = useState<UserTransaction[]>([])
  const [dailySummary, setDailySummary] = useState<DailySummaryPoint[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [datePreset, setDatePreset] = useState<DateRangePreset>('all')

  useEffect(() => {
    let cancelled = false
    const controller = new AbortController()

    const loadTransactions = async () => {
      setIsLoading(true)
      setError(null)
      try {
        const rangeQs = userDetailDateRangeQuery(datePreset)
        const summaryQuery = rangeQs ? `bucket=day&${rangeQs}` : 'bucket=day'
        const [transactionsResponse, summaryResponse] = await Promise.all([
          fetch(apiUserTransactionListUrl(user.id, rangeQs), { signal: controller.signal }),
          fetch(apiUserTransactionSummaryUrl(user.id, summaryQuery), { signal: controller.signal }),
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
  }, [user.id, datePreset])

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
          timeLabel: formatDateTime(point.bucketStart),
          createdAt: point.bucketStart,
        })),
    [dailySummary],
  )

  const totalPages = Math.max(1, Math.ceil(transactions.length / pageSize))
  const pagedTransactions = useMemo(() => {
    const start = (page - 1) * pageSize
    return transactions.slice(start, start + pageSize)
  }, [transactions, page])

  const distinctCalendarDayCount = useMemo(() => {
    const keys = new Set<string>()
    for (const tx of transactions) {
      const k = localCalendarDayKey(tx.createdAt)
      if (k) keys.add(k)
    }
    return keys.size
  }, [transactions])

  const tableSegments = useMemo(
    () => transactionRowsWithDayBreaks(pagedTransactions),
    [pagedTransactions],
  )

  return (
    <div className={`space-y-3 ${ui.stackGap}`}>
      <div className="card bg-base-100 shadow-md">
        <div className={`card-body ${ui.cardBodyPadding}`}>
          <div className="flex min-w-0 flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
            <h2 className={`card-title min-w-0 truncate ${ui.cardTitleText}`}>
              {user.name} {user.surname}
            </h2>
            <span className={`badge badge-error badge-outline shrink-0 ${ui.badgeSize}`}>
              Anomaly Rate: {anomalyRatio.toFixed(2)}%
            </span>
          </div>
          {error && <div className={`alert alert-error ${ui.tableCellText}`}>{error}</div>}
          {isLoading && <div className={`alert alert-info ${ui.tableCellText}`}>Loading user transactions...</div>}
          <div className={`${ui.chartHeight} min-w-0 w-full max-w-full`}>
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
                    if (!point) return 'Day: —'
                    return `Day: ${formatCalendarDayHeading(point.createdAt)}`
                  }}
                />
                <Bar yAxisId="amount" dataKey="amount" fill="#60a5fa" name="Daily amount sum" />
                <Line yAxisId="anomaly" type="monotone" dataKey="anomalyCount" stroke="#ef4444" strokeWidth={2} />
              </ComposedChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      <div className="card bg-base-100 shadow-md">
        <div className={`card-body ${ui.cardBodyPadding}`}>
          <div className="flex min-w-0 flex-col gap-1">
            <h3 className={`card-title ${ui.cardTitleText}`}>Transactions</h3>
            {!isLoading && transactions.length > 0 && distinctCalendarDayCount > 1 ? (
              <p className={`${ui.tableCellText} text-base-content/70`}>
                {distinctCalendarDayCount} calendar days with activity; table rows are grouped by day. The chart shows
                daily totals.
              </p>
            ) : null}
          </div>
          <div className="mb-3 flex min-w-0 flex-col gap-2 sm:flex-row sm:items-end sm:justify-end">
            <label className="form-control w-full max-w-[14rem] shrink-0">
              <span className={`label py-0 ${ui.tableCellText}`}>
                <span className="label-text font-medium">Date</span>
              </span>
              <select
                className={`select select-bordered w-full ${isMd ? 'select-xs' : 'select-sm'}`}
                value={datePreset}
                disabled={isLoading}
                onChange={(e) => setDatePreset(e.target.value as DateRangePreset)}
                aria-label="Filter transactions by date range"
              >
                {DATE_PRESET_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </label>
          </div>
          <div className={`min-w-0 max-w-full ${isMd ? 'max-h-[22rem]' : 'max-h-[28rem]'} overflow-x-auto overflow-y-auto`}>
            <table className={`table table-zebra ${ui.tableSize} min-w-[20rem]`}>
              <thead>
                <tr className={`sticky top-0 z-10 bg-base-100 ${ui.tableCellText}`}>
                  <th>ID</th>
                  <th>Amount</th>
                  {!isMd && <th>Location</th>}
                  <th>Anomaly</th>
                  <th>Date</th>
                </tr>
              </thead>
              <tbody>
                {!isLoading && pagedTransactions.length === 0 ? (
                  <tr>
                    <td colSpan={isMd ? 4 : 5} className="text-center opacity-70">
                      No transactions for this user.
                    </td>
                  </tr>
                ) : (
                  tableSegments.map((seg) =>
                    seg.kind === 'day' ? (
                      <tr key={`day-${seg.dayKey}`} className="bg-base-200/90">
                        <td
                          colSpan={isMd ? 4 : 5}
                          className={`${ui.tableCellText} py-2 text-sm font-semibold tracking-tight`}
                        >
                          {seg.label}
                        </td>
                      </tr>
                    ) : (
                      <tr key={seg.tx.id} className={ui.tableCellText}>
                        <td className="font-medium">{seg.tx.id}</td>
                        <td>${seg.tx.amount.toLocaleString()}</td>
                        {!isMd && <td>{seg.tx.location}</td>}
                        <td>
                          <span
                            className={`inline-flex h-6 w-[5.75rem] shrink-0 items-center justify-center rounded-[8px] text-xs font-medium text-white ${
                              seg.tx.fraudStatus ? 'bg-red-600' : 'bg-green-600'
                            }`}
                          >
                            {seg.tx.fraudStatus ? 'Anomaly' : 'Approved'}
                          </span>
                        </td>
                        <td className="whitespace-nowrap">{formatDateTime(seg.tx.createdAt)}</td>
                      </tr>
                    ),
                  )
                )}
              </tbody>
            </table>
          </div>
          {transactions.length > 0 && (
            <div className="mt-3 flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <span className={`min-w-0 ${ui.tableCellText}`}>
                Page {page} / {totalPages} - 100 rows per page
              </span>
              <div className="join shrink-0">
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
