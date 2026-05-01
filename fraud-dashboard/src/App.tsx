import { useState } from 'react'
import DashboardTable from './components/DashboardTable'
import Chart from './components/Chart'
import Sidebar from './components/Sidebar'
import UsersPage from './components/UsersPage'
import { useScreen } from './config/ScreenContext'
import { useLiveTransactions } from './hooks/useLiveTransactions'

function App() {
  const { isMd, ui } = useScreen()
  const [activePage, setActivePage] = useState<'live-data' | 'users'>('live-data')
  const env = import.meta.env as Record<string, string | undefined>
  const sseUrl = env.VITE_SSE_URL ?? '/stream/transactions'
  const usersUrl = env.VITE_USERS_URL ?? '/users'
  const userTransactionsBaseUrl = env.VITE_USER_TRANSACTIONS_URL ?? '/transactions/user'

  const { transactions, fraudRateSeries, chartTotals, connectionMode, isLoading, error } =
    useLiveTransactions({
      sseUrl,
      maxItems: 100,
    })

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
                <UsersPage usersUrl={usersUrl} transactionsBaseUrl={userTransactionsBaseUrl} />
              )}
            </div>
          </div>
        </div>
      </div>
    </main>
  )
}

export default App
