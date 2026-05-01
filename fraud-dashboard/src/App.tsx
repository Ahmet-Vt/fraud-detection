import { useMemo, useState } from 'react'
import DashboardTable from './components/DashboardTable'
import Chart from './components/Chart'
import Sidebar from './components/Sidebar'
import { useScreen } from './config/ScreenContext'
import { useLiveTransactions } from './hooks/useLiveTransactions'

function App() {
  const { isMd, ui } = useScreen()
  const [activePage, setActivePage] = useState<'live-data' | 'users'>('live-data')
  const env = import.meta.env as Record<string, string | undefined>
  const sseUrl = env.VITE_SSE_URL ?? '/stream/transactions'

  const { transactions, fraudRateSeries, chartTotals, connectionMode, isLoading, error } =
    useLiveTransactions({
      sseUrl,
      maxItems: 100,
    })

  const userRows = useMemo(() => {
    const grouped = new Map<
      string,
      { userLabel: string; transactionCount: number; fraudCount: number; latestSeenAt: string }
    >()

    for (const tx of transactions) {
      const userLabel = `${tx.userName} ${tx.userSurname}`.trim() || tx.userid
      const existing = grouped.get(tx.userid)
      if (!existing) {
        grouped.set(tx.userid, {
          userLabel,
          transactionCount: 1,
          fraudCount: tx.fraudstatus ? 1 : 0,
          latestSeenAt: tx.createdat,
        })
        continue
      }

      existing.transactionCount += 1
      if (tx.fraudstatus) existing.fraudCount += 1
      if (new Date(tx.createdat).getTime() > new Date(existing.latestSeenAt).getTime()) {
        existing.latestSeenAt = tx.createdat
      }
    }

    return [...grouped.entries()]
      .map(([userId, stats]) => ({
        userId,
        ...stats,
      }))
      .sort((a, b) => b.transactionCount - a.transactionCount)
  }, [transactions])

  return (
    <main className="min-h-screen bg-base-200">
      <div className="flex min-h-screen w-full">
        {!isMd && (
          <Sidebar
            isMobile={false}
            activePage={activePage}
            onNavigate={setActivePage}
            connectionMode={connectionMode}
            requestCount={chartTotals.requestCount}
            fraudCount={chartTotals.fraudCount}
          />
        )}

        <div className="flex-1">
          <div className={ui.pagePadding}>
            {isMd && (
              <Sidebar
                isMobile
                activePage={activePage}
                onNavigate={setActivePage}
                connectionMode={connectionMode}
                requestCount={chartTotals.requestCount}
                fraudCount={chartTotals.fraudCount}
              />
            )}
            <div className={`mx-auto ${ui.containerWidth} ${ui.stackGap}`}>
              <h1 className={`${ui.titleText} font-semibold`}>Fraud Dashboard</h1>

              {activePage === 'live-data' ? (
                <>
                  <Chart
                    data={fraudRateSeries}
                    requestCount={chartTotals.requestCount}
                    fraudCount={chartTotals.fraudCount}
                  />
                  <DashboardTable
                    rows={transactions}
                    connectionMode={connectionMode}
                    isLoading={isLoading}
                    error={error}
                  />
                </>
              ) : (
                <div className="card bg-base-100 shadow-md">
                  <div className={`card-body ${ui.cardBodyPadding}`}>
                    <div className="flex items-center justify-between gap-2">
                      <h2 className={`card-title ${ui.cardTitleText}`}>Users</h2>
                      <span className={`badge badge-outline ${ui.badgeSize}`}>{userRows.length} users</span>
                    </div>
                    <div className={`${isMd ? 'max-h-[22rem]' : 'max-h-[28rem]'} overflow-auto`}>
                      <table className={`table table-zebra ${ui.tableSize}`}>
                        <thead>
                          <tr className={`sticky top-0 z-10 bg-base-100 ${ui.tableCellText}`}>
                            <th>User</th>
                            <th>User ID</th>
                            <th>Transactions</th>
                            <th>Fraud</th>
                            {!isMd && <th>Last Seen</th>}
                          </tr>
                        </thead>
                        <tbody>
                          {userRows.length === 0 ? (
                            <tr>
                              <td colSpan={isMd ? 4 : 5} className="text-center opacity-70">
                                No user activity yet.
                              </td>
                            </tr>
                          ) : (
                            userRows.map((row) => (
                              <tr key={row.userId} className={ui.tableCellText}>
                                <td>{row.userLabel}</td>
                                <td className="font-medium">{row.userId}</td>
                                <td>{row.transactionCount}</td>
                                <td>{row.fraudCount}</td>
                                {!isMd && <td>{row.latestSeenAt}</td>}
                              </tr>
                            ))
                          )}
                        </tbody>
                      </table>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </main>
  )
}

export default App
