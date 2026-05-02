import { useCallback, useState } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import { apiSseLiveTransactionsUrl } from '../api/endpoints'
import { useScreen } from '../config/ScreenContext'
import { useLiveTransactions, type LiveTransaction } from '../features/live/useLiveTransactions'
import AnomalyInboxFab, { type AnomalyInboxItem } from './AnomalyInboxFab'
import Sidebar from './Sidebar'

const liveTransactionMaxItems = 100
const anomalyInboxMaxItems = 300

function anomalyInboxUserLabel(tx: LiveTransaction): string {
  return [tx.userName, tx.userSurname].filter(Boolean).join(' ').trim() || 'unknown user'
}

export default function DashboardLayout() {
  const { isMd, ui } = useScreen()
  const navigate = useNavigate()
  const [anomalyItems, setAnomalyItems] = useState<AnomalyInboxItem[]>([])

  const onFraudTransaction = useCallback((tx: LiveTransaction) => {
    setAnomalyItems((prev) =>
      [{ id: tx.id, userId: tx.userid, userLabel: anomalyInboxUserLabel(tx) }, ...prev].slice(
        0,
        anomalyInboxMaxItems,
      ),
    )
  }, [])

  const onAnomalyRowClick = useCallback(
    (item: AnomalyInboxItem) => {
      navigate(`/users/${encodeURIComponent(item.userId)}`)
    },
    [navigate],
  )

  const live = useLiveTransactions({
    sseUrl: apiSseLiveTransactionsUrl,
    maxItems: liveTransactionMaxItems,
    onFraudTransaction,
  })

  return (
    <main className="relative min-h-screen overflow-x-hidden bg-base-200">
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
      <AnomalyInboxFab items={anomalyItems} onRowClick={onAnomalyRowClick} />
    </main>
  )
}
