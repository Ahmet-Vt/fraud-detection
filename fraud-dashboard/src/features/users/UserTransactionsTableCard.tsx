import { useEffect, useMemo, useState } from 'react'
import { useScreen } from '../../config/ScreenContext'
import { formatCalendarDayHeading, formatDateTime, localCalendarDayKey } from '../../lib/formatDateTime'
import { AppCard, FraudStatusBadge, SimplePagination, tableScrollAreaClass } from '../../components/ui/dashboardUi'
import type { UserTransaction, UserTransactionDateRange } from './userDetailTransactions'
import UserTransactionRangePickerButton from './UserTransactionRangePickerButton'

const pageSize = 100

type TransactionTableSegment =
  | { kind: 'day'; dayKey: string; label: string }
  | { kind: 'row'; tx: UserTransaction }

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

type UserTransactionsTableCardProps = {
  transactions: UserTransaction[]
  dateRange: UserTransactionDateRange
  onDateRangeChange: (range: UserTransactionDateRange) => void
  isLoading: boolean
}

export default function UserTransactionsTableCard({
  transactions,
  dateRange,
  onDateRangeChange,
  isLoading,
}: UserTransactionsTableCardProps) {
  const { isMd, ui } = useScreen()
  const [page, setPage] = useState(1)

  useEffect(() => {
    setPage(1)
  }, [transactions])

  const totalPages = Math.max(1, Math.ceil(transactions.length / pageSize))
  const pagedTransactions = useMemo(() => {
    const start = (page - 1) * pageSize
    return transactions.slice(start, start + pageSize)
  }, [transactions, page])

  const tableSegments = useMemo(
    () => transactionRowsWithDayBreaks(pagedTransactions),
    [pagedTransactions],
  )

  return (
    <AppCard bodyPaddingClass={ui.cardBodyPadding}>
      <h3 className={`card-title ${ui.cardTitleText}`}>Transactions</h3>
      <div className="mb-3 flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <p className={`m-0 min-w-0 opacity-80 ${ui.tableCellText}`}>Filter by date range</p>
        <UserTransactionRangePickerButton
          dateRange={dateRange}
          onDateRangeChange={onDateRangeChange}
          isLoading={isLoading}
          buttonSizeClass={isMd ? 'btn-sm' : ''}
          textClass={ui.tableCellText}
        />
      </div>
      <div className={tableScrollAreaClass(isMd)}>
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
                      <FraudStatusBadge isFraud={seg.tx.fraudStatus} variant="userDetailPill" />
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
        <SimplePagination
          page={page}
          totalPages={totalPages}
          isMd={isMd}
          textClass={ui.tableCellText}
          pageSize={pageSize}
          onPrev={() => setPage((p) => Math.max(1, p - 1))}
          onNext={() => setPage((p) => Math.min(totalPages, p + 1))}
        />
      )}
    </AppCard>
  )
}
