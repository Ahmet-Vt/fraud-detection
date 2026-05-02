import { Outlet } from 'react-router-dom'
import { apiSseLiveTransactionsUrl } from '../api/endpoints'
import Sidebar from '../components/Sidebar'
import { useScreen } from '../config/ScreenContext'
import { useLiveTransactions } from '../hooks/useLiveTransactions'

export default function DashboardLayout() {
  const { isMd, ui } = useScreen()
  const live = useLiveTransactions({
    sseUrl: apiSseLiveTransactionsUrl,
    maxItems: 100,
  })

  return (
    <main className="min-h-screen overflow-x-hidden bg-base-200">
      <div className="flex min-h-screen w-full min-w-0">
        {!isMd && (
          <Sidebar
            isMobile={false}
            connectionMode={live.connectionMode}
            requestCount={live.chartTotals.requestCount}
            fraudCount={live.chartTotals.fraudCount}
          />
        )}

        <div className="flex min-h-0 min-w-0 flex-1 flex-col">
          <div className={`${ui.pagePadding} min-w-0`}>
            {isMd ? (
              <div className="mb-3 flex min-w-0 items-center gap-2">
                <Sidebar
                  isMobile
                  connectionMode={live.connectionMode}
                  requestCount={live.chartTotals.requestCount}
                  fraudCount={live.chartTotals.fraudCount}
                />
                <h1 className={`${ui.titleText} min-w-0 flex-1 truncate font-semibold`}>Fraud Dashboard</h1>
              </div>
            ) : null}
            <div className={`mx-auto min-w-0 ${ui.containerWidth} ${ui.stackGap}`}>
              {!isMd ? <h1 className={`${ui.titleText} font-semibold`}>Fraud Dashboard</h1> : null}
              <Outlet context={live} />
            </div>
          </div>
        </div>
      </div>
    </main>
  )
}
