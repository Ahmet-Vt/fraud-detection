import Chart from './Chart'
import DashboardTable from './DashboardTable'
import type { LiveTransactionsOutletContext } from '../hooks/useLiveTransactions'

export default function LiveDataPanel({
  transactions,
  fraudRateSeries,
  chartTotals,
  connectionMode,
  isLoading,
  error,
}: LiveTransactionsOutletContext) {
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
