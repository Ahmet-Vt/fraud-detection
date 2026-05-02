import { Outlet } from 'react-router-dom'
import Sidebar from '../components/Sidebar'
import { useScreen } from '../config/ScreenContext'
import { useLiveTransactions } from '../hooks/useLiveTransactions'

export default function DashboardLayout() {
  const { isMd, ui } = useScreen()
  const env = import.meta.env as Record<string, string | undefined>
  const sseUrl = env.VITE_SSE_URL ?? '/stream/transactions'

  const live = useLiveTransactions({
    sseUrl,
    maxItems: 100,
  })

  return (
    <main className="min-h-screen bg-base-200">
      <div className="flex min-h-screen w-full">
        {!isMd && (
          <Sidebar
            isMobile={false}
            connectionMode={live.connectionMode}
            requestCount={live.chartTotals.requestCount}
            fraudCount={live.chartTotals.fraudCount}
          />
        )}

        <div className="flex-1">
          <div className={ui.pagePadding}>
            {isMd && (
              <Sidebar
                isMobile
                connectionMode={live.connectionMode}
                requestCount={live.chartTotals.requestCount}
                fraudCount={live.chartTotals.fraudCount}
              />
            )}
            <div className={`mx-auto ${ui.containerWidth} ${ui.stackGap}`}>
              <h1 className={`${ui.titleText} font-semibold`}>Fraud Dashboard</h1>
              <Outlet context={live} />
            </div>
          </div>
        </div>
      </div>
    </main>
  )
}
