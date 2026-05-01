import DashboardTable from './components/DashboardTable'
import Chart from './components/Chart'
import { useScreen } from './config/ScreenContext'

function App() {
  const { ui } = useScreen()

  return (
    <main className={`min-h-screen bg-base-200 ${ui.pagePadding}`}>
      <div className={`mx-auto ${ui.containerWidth} ${ui.stackGap}`}>
        <h1 className={`${ui.titleText} font-semibold`}>Fraud Dashboard</h1>
        <Chart />
        <DashboardTable />
      </div>
    </main>
  )
}

export default App
