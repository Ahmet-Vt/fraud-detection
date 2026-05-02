import { useOutletContext } from 'react-router-dom'
import Chart from '../components/Chart'
import DashboardTable from '../components/DashboardTable'
import type { LiveTransactionsOutletContext } from '../hooks/useLiveTransactions'

export default function LiveDataPage() {
  const {
    transactions,
    fraudRateSeries,
    chartTotals,
    connectionMode,
    isLoading,
    error,
  } = useOutletContext<LiveTransactionsOutletContext>()

  return (
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
  )
}
