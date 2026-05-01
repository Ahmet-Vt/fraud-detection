import { useLayoutEffect, useRef } from 'react'
import { useScreen } from '../config/ScreenContext'
import type { LiveTransaction } from '../hooks/useLiveTransactions'

type DashboardTableProps = {
  rows: LiveTransaction[]
  connectionMode: 'idle' | 'sse' | 'disconnected'
  isLoading?: boolean
  error?: string | null
}

export default function DashboardTable({
  rows,
  connectionMode,
  isLoading = false,
  error = null,
}: DashboardTableProps) {
  const { isMd, ui } = useScreen()
  const scrollContainerRef = useRef<HTMLDivElement | null>(null)
  const prevScrollStateRef = useRef<{ top: number; height: number }>({ top: 0, height: 0 })
  const connectionBadgeClass =
    connectionMode === 'sse'
      ? 'badge badge-success badge-outline'
      : connectionMode === 'disconnected'
        ? 'badge badge-warning badge-outline'
        : 'badge badge-ghost'

  useLayoutEffect(() => {
    const el = scrollContainerRef.current
    if (!el) return

    const prev = prevScrollStateRef.current
    const nextHeight = el.scrollHeight
    const heightDelta = nextHeight - prev.height
    const wasNearTop = prev.top <= 8

    if (prev.height > 0 && heightDelta > 0 && !wasNearTop) {
      el.scrollTop = prev.top + heightDelta
    }

    prevScrollStateRef.current = { top: el.scrollTop, height: el.scrollHeight }
  }, [rows])

  const handleScroll = () => {
    const el = scrollContainerRef.current
    if (!el) return
    prevScrollStateRef.current.top = el.scrollTop
  }

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

        {error && <div className={`alert alert-error ${ui.tableCellText}`}>{error}</div>}
        {isLoading && <div className={`alert alert-info ${ui.tableCellText}`}>Loading live data...</div>}

        <div
          ref={scrollContainerRef}
          onScroll={handleScroll}
          className={`${isMd ? 'max-h-[22rem]' : 'max-h-[28rem]'} overflow-auto [overflow-anchor:none]`}
        >
          <table className={`table table-zebra ${ui.tableSize}`}>
            <thead>
              <tr className={`sticky top-0 z-10 bg-base-100 ${ui.tableCellText}`}>
                <th>User</th>
                <th>Transaction</th>
                <th>Amount</th>
                {!isMd && <th>Location</th>}
                <th>Fraud</th>
              </tr>
            </thead>
            <tbody>
              {rows.length === 0 ? (
                <tr>
                  <td colSpan={isMd ? 4 : 5} className="text-center opacity-70">
                    No transaction received yet.
                  </td>
                </tr>
              ) : (
                rows.map((row) => (
                  <tr
                    key={row.id}
                    className={`${ui.tableCellText} ${row.fraudstatus ? 'bg-red-50/80 hover:bg-red-100/80' : ''}`}
                  >
                    <td>{`${row.userName} ${row.userSurname}`.trim() || row.userid}</td>
                    <td className="font-medium">{row.id}</td>
                    <td>${row.amount.toLocaleString()}</td>
                    {!isMd && <td>{row.location}</td>}
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
