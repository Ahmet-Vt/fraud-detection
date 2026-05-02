import { memo } from 'react'
import Chart from './Chart'
import DashboardTable from './DashboardTable'
import type { LiveTransactionsOutletContext } from '../hooks/useLiveTransactions'

function LiveDataPanel({
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

export default memo(LiveDataPanel)
