import DashboardTable from './components/DashboardTable'
import Chart from './components/Chart'
import { useScreen } from './config/ScreenContext'
import { useLiveTransactions } from './hooks/useLiveTransactions'

function App() {
  const { ui } = useScreen()
  const env = import.meta.env as Record<string, string | undefined>
  const sseUrl = env.VITE_SSE_URL ?? '/stream/transactions'

  const { transactions, connectionMode, isLoading, error } = useLiveTransactions({
    sseUrl,
    maxItems: 100,
  })

  return (
    <main className={`min-h-screen bg-base-200 ${ui.pagePadding}`}>
      <div className={`mx-auto ${ui.containerWidth} ${ui.stackGap}`}>
        <h1 className={`${ui.titleText} font-semibold`}>Fraud Dashboard</h1>
        <Chart />
        <DashboardTable
          rows={transactions}
          connectionMode={connectionMode}
          isLoading={isLoading}
          error={error}
        />
      </div>
    </main>
  )
}

export default App
