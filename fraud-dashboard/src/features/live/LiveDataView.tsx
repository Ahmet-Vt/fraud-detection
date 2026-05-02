import { memo, useCallback, useLayoutEffect, useMemo, useRef } from 'react'
import { CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { useScreen } from '../../config/ScreenContext'
import { formatLiveDateTime } from '../../lib/formatDateTime'
import { AppCard, FraudStatusBadge, InlineAlert, tableScrollAreaClass } from '../../components/ui/dashboardUi'
import type { FraudRatePoint, LiveTransaction, LiveTransactionsOutletContext } from './useLiveTransactions'

type LiveAnomalyRateChartProps = {
  data: FraudRatePoint[]
  requestCount: number
  fraudCount: number
}

function LiveAnomalyRateChart({ data, requestCount, fraudCount }: LiveAnomalyRateChartProps) {
  const { isXs, isMd, isSm, ui } = useScreen()
  const chartData = useMemo(() => data.map((point, index) => ({ ...point, pointIndex: index })), [data])
  const xAxisTicks = useMemo(() => {
    if (chartData.length === 0) return []
    const ticks: number[] = []
    for (let i = 0; i < chartData.length; i += 60) {
      ticks.push(i)
    }
    const lastTick = chartData.length - 1
    if (ticks[ticks.length - 1] !== lastTick) {
      ticks.push(lastTick)
    }
    return ticks
  }, [chartData])

  return (
    <AppCard bodyPaddingClass={ui.cardBodyPadding}>
      <div className={`flex justify-between gap-2 ${isSm ? 'flex-col items-start' : 'items-center'}`}>
        <h2 className={`card-title ${ui.cardTitleText}`}>1-Hour Anomaly Rate</h2>
        <div className={`flex items-center gap-2 ${isSm ? 'w-full flex-wrap justify-start' : 'flex-nowrap'}`}>
          <span className={`badge badge-outline ${ui.badgeSize}`}>Per 15 Seconds</span>
          <span className={`badge badge-info badge-outline ${ui.badgeSize}`}>
            {isSm ? `Req: ${requestCount}` : `Requests: ${requestCount}`}
          </span>
          <span className={`badge badge-error badge-outline ${ui.badgeSize}`}>Fraud: {fraudCount}</span>
        </div>
      </div>

      <div className={`${ui.chartHeight} min-w-0 w-full max-w-full`}>
        <ResponsiveContainer width="100%" height="100%">
          <LineChart
            data={chartData}
            margin={{
              top: 8,
              right: isSm ? 4 : 8,
              bottom: isSm ? 0 : 8,
              left: isSm ? 2 : 8,
            }}
          >
            {ui.chartGridVisible && <CartesianGrid strokeDasharray="3 3" />}
            <XAxis
              dataKey="pointIndex"
              type="number"
              ticks={xAxisTicks}
              tickFormatter={(value: number) => chartData[value]?.minuteLabel ?? ''}
              minTickGap={isSm ? 24 : isMd ? 20 : 32}
            />
            <YAxis
              unit="%"
              domain={[0, 100]}
              ticks={[0, 25, 50, 75, 100]}
              allowDecimals={false}
              width={isXs ? 38 : isSm ? 44 : isMd ? 48 : 52}
            />
            <Tooltip
              filterNull={false}
              labelFormatter={(value) => {
                const index = Number(value)
                const point = chartData[index]
                if (!point) return 'Time: --:--'
                return `Time: ${point.minuteLabel} | Requests: ${point.requestCount} | Fraud: ${point.fraudCount}`
              }}
              formatter={(value) =>
                value == null ? ['No data yet', 'Anomaly Rate'] : [`${Number(value).toFixed(2)}%`, 'Anomaly Rate']
              }
            />
            <Line
              type="monotone"
              dataKey="anomalyRate"
              connectNulls
              stroke="#ef4444"
              strokeWidth={2}
              dot={{ r: ui.chartDotR }}
              activeDot={{ r: ui.chartActiveDotR }}
              name="Anomaly Rate"
              isAnimationActive={false}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </AppCard>
  )
}

type LiveTransactionFeedTableProps = {
  rows: LiveTransaction[]
  connectionMode: 'idle' | 'sse' | 'disconnected'
  isLoading?: boolean
  error?: string | null
}

function LiveTransactionFeedTable({
  rows,
  connectionMode,
  isLoading = false,
  error = null,
}: LiveTransactionFeedTableProps) {
  const { isMd, ui } = useScreen()
  const scrollContainerRef = useRef<HTMLDivElement | null>(null)
  const prevScrollStateRef = useRef<{ top: number; height: number }>({ top: 0, height: 0 })
  const connectionBadgeClass =
    connectionMode === 'sse'
      ? 'badge badge-success badge-outline'
      : connectionMode === 'disconnected'
        ? 'badge badge-warning badge-outline'
        : 'badge badge-ghost'

  useLayoutEffect(() => {
    const el = scrollContainerRef.current
    if (!el) return

    const prev = prevScrollStateRef.current
    const nextHeight = el.scrollHeight
    const heightDelta = nextHeight - prev.height
    const wasNearTop = prev.top <= 8

    if (prev.height > 0 && heightDelta > 0 && !wasNearTop) {
      el.scrollTop = prev.top + heightDelta
    }

    prevScrollStateRef.current = { top: el.scrollTop, height: el.scrollHeight }
  }, [rows])

  const handleScroll = useCallback(() => {
    const el = scrollContainerRef.current
    if (!el) return
    prevScrollStateRef.current = {
      ...prevScrollStateRef.current,
      top: el.scrollTop,
    }
  }, [])

  return (
    <AppCard bodyPaddingClass={ui.cardBodyPadding}>
      <div className="flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <h2 className={`card-title min-w-0 ${ui.cardTitleText}`}>Live Transaction Feed</h2>
        <div className="flex shrink-0 flex-wrap items-center gap-2">
          <span className={`${connectionBadgeClass} ${ui.badgeSize}`}>{connectionMode}</span>
          <span className={`badge badge-info badge-outline ${ui.badgeSize}`}>{rows.length} items</span>
        </div>
      </div>

      {error && <InlineAlert variant="error" textClass={ui.tableCellText}>{error}</InlineAlert>}
      {isLoading && <InlineAlert variant="info" textClass={ui.tableCellText}>Loading live data...</InlineAlert>}

      <div
        ref={scrollContainerRef}
        onScroll={handleScroll}
        className={tableScrollAreaClass(isMd, { overflowAnchorNone: true })}
      >
        <table className={`table table-zebra ${ui.tableSize} min-w-[18rem]`}>
          <thead>
            <tr className={`sticky top-0 z-10 bg-base-100 ${ui.tableCellText}`}>
              <th>User</th>
              <th>Transaction</th>
              <th>Amount</th>
              {!isMd && <th>Location</th>}
              <th>Fraud</th>
              <th>Created</th>
            </tr>
          </thead>
          <tbody>
            {rows.length === 0 ? (
              <tr>
                <td colSpan={isMd ? 5 : 6} className="text-center opacity-70">
                  No transaction received yet.
                </td>
              </tr>
            ) : (
              rows.map((row) => (
                <tr
                  key={row.id}
                  className={`${ui.tableCellText} ${row.fraudstatus ? 'bg-red-50/80 hover:bg-red-100/80' : ''}`}
                >
                  <td>{`${row.userName} ${row.userSurname}`.trim() || row.userid}</td>
                  <td className="font-medium">{row.id}</td>
                  <td>${row.amount.toLocaleString()}</td>
                  {!isMd && <td>{row.location}</td>}
                  <td>
                    <FraudStatusBadge
                      isFraud={row.fraudstatus}
                      variant="liveTable"
                      badgeSizeClass={ui.badgeSize}
                    />
                  </td>
                  <td className="whitespace-nowrap">{formatLiveDateTime(row.createdat)}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </AppCard>
  )
}

function LiveDataView({
  transactions,
  fraudRateSeries,
  chartTotals,
  connectionMode,
  isLoading,
  error,
}: LiveTransactionsOutletContext) {
  return (
    <>
      <LiveAnomalyRateChart
        data={fraudRateSeries}
        requestCount={chartTotals.requestCount}
        fraudCount={chartTotals.fraudCount}
      />
      <LiveTransactionFeedTable
        rows={transactions}
        connectionMode={connectionMode}
        isLoading={isLoading}
        error={error}
      />
    </>
  )
}

export default memo(LiveDataView)
