import { useEffect, useMemo, useState } from 'react'
import { useScreen } from '../../config/ScreenContext'
import { formatCalendarDayHeading, formatDateTime, localCalendarDayKey } from '../../lib/formatDateTime'
import { AppCard, FraudStatusBadge, SimplePagination, tableScrollAreaClass } from '../../components/ui/dashboardUi'
import type { UserTransaction, UserTransactionDateRange } from './userDetailTransactions'
import UserTransactionRangePickerButton from './UserTransactionRangePickerButton'

const FRAUD_RULE_LABEL: Record<string, string> = {
  amount: 'Amount',
  location: 'Location',
  velocity: 'Velocity',
}

function fraudRuleLabel(ruleName: string): string {
  return FRAUD_RULE_LABEL[ruleName] ?? ruleName
}

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
  const [hoveredTx, setHoveredTx] = useState<UserTransaction | null>(null)

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
      <div className="flex min-w-0 flex-col gap-3 lg:flex-row lg:items-stretch">
        <div className={['min-w-0 flex-1', tableScrollAreaClass(isMd)].join(' ')}>
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
                    <tr
                      key={seg.tx.id}
                      className={`${ui.tableCellText} cursor-default transition-colors hover:bg-base-200/60`}
                      onMouseEnter={() => setHoveredTx(seg.tx)}
                      onMouseLeave={() => setHoveredTx(null)}
                    >
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
        <aside className="w-full shrink-0 lg:sticky lg:top-4 lg:self-start lg:w-72">
          <AppCard bodyPaddingClass={`${ui.cardBodyPadding} py-4`} className="shadow-sm">
            <h4 className={`card-title mb-2 text-sm ${ui.cardTitleText}`}>Fraud rules</h4>
            {!hoveredTx ? (
              <p className={`m-0 text-sm opacity-70 ${ui.tableCellText}`}>
                Hover a transaction row to see how each rule evaluated for that request.
              </p>
            ) : (
              <div className={`flex flex-col gap-3 ${ui.tableCellText}`}>
                <div className="min-w-0">
                  <p className="m-0 truncate text-xs opacity-60">Transaction</p>
                  <p className="m-0 truncate font-mono text-xs font-medium" title={hoveredTx.id}>
                    {hoveredTx.id}
                  </p>
                  <p className="mt-1 m-0 text-xs opacity-80">
                    Status: <span className="font-medium capitalize">{hoveredTx.transactionStatus}</span>
                  </p>
                </div>
                {hoveredTx.transactionStatus === 'pending' ? (
                  <p className="m-0 text-sm opacity-80">Fraud evaluation has not finished for this transaction yet.</p>
                ) : hoveredTx.fraudRules.length === 0 ? (
                  <p className="m-0 text-sm opacity-80">No stored rule results for this transaction.</p>
                ) : (
                  <ul className="m-0 list-none space-y-2 p-0">
                    {hoveredTx.fraudRules.map((r) => (
                      <li
                        key={`${hoveredTx.id}-${r.ruleName}`}
                        className="flex items-center justify-between gap-2 rounded-lg bg-base-200/50 px-2 py-1.5 text-sm"
                      >
                        <span className="min-w-0 font-medium">{fraudRuleLabel(r.ruleName)}</span>
                        <span
                          className={
                            r.matched
                              ? `badge shrink-0 ${ui.badgeSize} badge-error badge-outline`
                              : `badge shrink-0 ${ui.badgeSize} badge-success badge-outline`
                          }
                        >
                          {r.matched ? 'Matched' : 'Clear'}
                        </span>
                      </li>
                    ))}
                  </ul>
                )}
                <p className="m-0 border-t border-base-300 pt-2 text-xs opacity-60">
                  Anomaly flag uses two or more matched rules. Single rule hits stay &quot;clear&quot; at account level.
                </p>
              </div>
            )}
          </AppCard>
        </aside>
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
