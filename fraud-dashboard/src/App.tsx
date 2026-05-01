import DashboardTable from './components/DashboardTable'
import Chart from './components/Chart'

function App() {
  return (
    <main className="min-h-screen bg-base-200 p-4 md:p-8">
      <div className="mx-auto max-w-6xl space-y-4">
        <h1 className="text-2xl font-semibold">Fraud Dashboard</h1>
        <Chart />
        <DashboardTable />
      </div>
    </main>
  )
}

export default App
