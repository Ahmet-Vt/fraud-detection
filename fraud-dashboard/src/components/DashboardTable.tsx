import { useScreen } from '../config/ScreenContext'
import type { LiveTransaction } from '../hooks/useLiveTransactions'

type DashboardTableProps = {
  rows: LiveTransaction[]
  connectionMode: 'idle' | 'sse' | 'disconnected'
  isLoading?: boolean
  error?: string | null
}

function statusClass(status: string) {
  if (status === 'completed') return 'badge badge-success badge-soft'
  if (status === 'failed') return 'badge badge-error badge-soft'
  return 'badge badge-warning badge-soft'
}

export default function DashboardTable({
  rows,
  connectionMode,
  isLoading = false,
  error = null,
}: DashboardTableProps) {
  const { isMd, ui } = useScreen()
  const connectionBadgeClass =
    connectionMode === 'sse'
      ? 'badge badge-success badge-outline'
      : connectionMode === 'disconnected'
        ? 'badge badge-warning badge-outline'
        : 'badge badge-ghost'

  return (
    <div className="card bg-base-100 shadow-md">
      <div className={`card-body ${ui.cardBodyPadding}`}>
        <div className="flex items-center justify-between gap-2">
          <h2 className={`card-title ${ui.cardTitleText}`}>Live Transaction Feed</h2>
          <div className="flex items-center gap-2">
            <span className={`${connectionBadgeClass} ${ui.badgeSize}`}>{connectionMode}</span>
            <span className={`badge badge-info badge-outline ${ui.badgeSize}`}>{rows.length} items</span>
          </div>
        </div>

        {error && <div className="alert alert-error text-sm">{error}</div>}
        {isLoading && <div className="alert alert-info text-sm">Loading live data...</div>}

        <div className="max-h-[28rem] overflow-auto">
          <table className={`table table-zebra ${ui.tableSize}`}>
            <thead>
              <tr className={`sticky top-0 z-10 bg-base-100 ${ui.tableCellText}`}>
                <th>Transaction</th>
                <th>User</th>
                <th>Amount</th>
                {!isMd && <th>Location</th>}
                <th>Status</th>
                <th>Fraud</th>
              </tr>
            </thead>
            <tbody>
              {rows.length === 0 ? (
                <tr>
                  <td colSpan={isMd ? 5 : 6} className="text-center opacity-70">
                    No transaction received yet.
                  </td>
                </tr>
              ) : (
                rows.map((row) => (
                  <tr
                    key={row.id}
                    className={`${ui.tableCellText} ${row.fraudstatus ? 'bg-red-50/80 hover:bg-red-100/80' : ''}`}
                  >
                    <td className="font-medium">{row.id}</td>
                    <td>{`${row.userName} ${row.userSurname}`.trim() || row.userid}</td>
                    <td>${row.amount.toLocaleString()}</td>
                    {!isMd && <td>{row.location}</td>}
                    <td>
                      <span className={`${statusClass(row.transactionStatus)} ${ui.badgeSize}`}>
                        {row.transactionStatus}
                      </span>
                    </td>
                    <td>
                      <span
                        className={`${row.fraudstatus ? 'badge bg-red-600 text-white border-red-700' : 'badge badge-success badge-soft'} ${ui.badgeSize}`}
                      >
                        {row.fraudstatus ? 'suspicious' : 'approved'}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
