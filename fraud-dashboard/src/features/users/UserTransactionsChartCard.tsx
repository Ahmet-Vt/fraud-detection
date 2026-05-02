import { useLayoutEffect, useMemo, useRef, useState } from 'react'
import { Bar, CartesianGrid, Line, ComposedChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { useScreen } from '../../config/ScreenContext'
import { formatCalendarDayHeading, formatChartBucketAxis } from '../../lib/formatDateTime'
import { formatUserRiskRatePercent } from '../../lib/formatRiskRate'
import { AppCard, InlineAlert } from '../../components/ui/dashboardUi'
import { pickChartBucketForRange, type DailySummaryPoint, type UserTransactionDateRange } from './userDetailTransactions'
import type { UserRow } from './userTypes'

const userChartColorAmount = '#3b82f6'
const userChartColorApproved = '#22c55e'
const userChartColorAnomaly = '#ef4444'

type UserTransactionsChartCardProps = {
  user: UserRow
  dailySummary: DailySummaryPoint[]
  dateRange: UserTransactionDateRange
  anomalyRatio: number
  isLoading: boolean
  error: string | null
}

export default function UserTransactionsChartCard({
  user,
  dailySummary,
  dateRange,
  anomalyRatio,
  isLoading,
  error,
}: UserTransactionsChartCardProps) {
  const { isXs, isSm, isMd, isDesktop, ui } = useScreen()

  const chartBucket = useMemo(() => pickChartBucketForRange(dateRange), [dateRange])

  const userChartResponsive = useMemo(() => {
    if (isXs) {
      return { chartInsetPercent: 3, maxBarSize: 20 }
    }
    if (isSm) {
      return { chartInsetPercent: 4, maxBarSize: 26 }
    }
    if (isMd) {
      return { chartInsetPercent: 5, maxBarSize: 30 }
    }
    return { chartInsetPercent: 5, maxBarSize: 36 }
  }, [isXs, isSm, isMd])

  const chartData = useMemo(
    () =>
      dailySummary.map((point) => ({
        amount: point.amountSum,
        approvedCount: Math.max(0, point.requestCount - point.anomalyCount),
        anomalyCount: point.anomalyCount,
        createdAt: point.bucketStart,
      })),
    [dailySummary],
  )

  const chartViewportRef = useRef<HTMLDivElement>(null)
  const [viewportWidth, setViewportWidth] = useState(0)

  useLayoutEffect(() => {
    const el = chartViewportRef.current
    if (!el || typeof ResizeObserver === 'undefined') return
    const ro = new ResizeObserver(() => setViewportWidth(el.clientWidth))
    ro.observe(el)
    setViewportWidth(el.clientWidth)
    return () => ro.disconnect()
  }, [])

  const chartPlotWidth = useMemo(() => {
    const n = chartData.length
    if (n <= 0) return undefined
    const minPxPerBucket = 12
    const minW = n * minPxPerBucket
    if (viewportWidth <= 0) return minW
    return Math.max(viewportWidth, minW)
  }, [chartData.length, viewportWidth])

  return (
    <AppCard bodyPaddingClass={ui.cardBodyPadding}>
      <div className="flex min-w-0 flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
        <h2 className="card-title m-0 min-w-0 max-w-full p-0">
          <div
            className={`min-w-0 max-w-full truncate rounded-[8px] px-3 py-2 outline outline-1 -outline-offset-1 outline-base-300 ${ui.cardTitleText}`}
          >
            <span className="font-normal text-base-content/70">Username:</span>{' '}
            <span className="font-semibold">
              {user.name} {user.surname}
            </span>
            <span className="inline-block pl-2 sm:pl-3">
              <span className="font-normal text-base-content/70">Risk rate:</span>{' '}
              <span className="font-semibold tabular-nums">
                {formatUserRiskRatePercent(user.riskRate, 2)}
              </span>
            </span>
          </div>
        </h2>
        <span className={`badge badge-error badge-outline shrink-0 ${ui.badgeSize}`}>
          Anomaly Rate: {anomalyRatio.toFixed(2)}%
        </span>
      </div>
      {error && <InlineAlert variant="error" textClass={ui.tableCellText}>{error}</InlineAlert>}
      {isLoading && (
        <InlineAlert variant="info" textClass={ui.tableCellText}>
          Loading user transactions...
        </InlineAlert>
      )}
      <div
        className={`flex min-w-0 w-full max-w-full flex-col gap-2 lg:flex-row lg:items-stretch lg:gap-3 ${isDesktop ? `${ui.chartHeight} min-h-0` : ''}`}
      >
        <div
          className={`box-border min-h-0 min-w-0 w-full lg:flex-1 lg:min-h-0 ${ui.chartHeight} ${isDesktop ? 'lg:h-full' : ''}`}
          style={{
            paddingLeft: `${userChartResponsive.chartInsetPercent}%`,
            paddingRight: `${userChartResponsive.chartInsetPercent}%`,
          }}
        >
          <div ref={chartViewportRef} className="h-full min-h-0 w-full overflow-x-auto">
            <div
              className="h-full min-h-0"
              style={chartPlotWidth != null ? { width: chartPlotWidth, minWidth: '100%' } : { width: '100%' }}
            >
              <ResponsiveContainer width="100%" height="100%">
                <ComposedChart
                  data={chartData}
                  margin={{
                    top: 8,
                    right: isDesktop ? (isSm ? 6 : 8) : isXs ? 4 : 6,
                    bottom: isSm ? 0 : 8,
                    left: isXs ? 10 : isSm ? 12 : 14,
                  }}
                >
                  {ui.chartGridVisible && <CartesianGrid strokeDasharray="3 3" />}
                  <XAxis
                    dataKey="createdAt"
                    type="category"
                    tickFormatter={(value: string) => formatChartBucketAxis(value, chartBucket)}
                    minTickGap={isSm ? 14 : isMd ? 20 : 28}
                    padding={{ left: 8, right: 8 }}
                  />
                  <YAxis yAxisId="amount" width={isXs ? 56 : isSm ? 62 : isMd ? 68 : 76} />
                  <YAxis
                    yAxisId="count"
                    orientation="right"
                    width={isXs ? 34 : isSm ? 38 : isMd ? 42 : 48}
                    allowDecimals={false}
                  />
                  <Tooltip
                    labelFormatter={(value: string | number) => {
                      const key = typeof value === 'string' ? value : chartData[Number(value)]?.createdAt
                      const point = key ? chartData.find((d) => d.createdAt === key) : undefined
                      if (!point) return '—'
                      if (chartBucket === 'hour') {
                        return formatChartBucketAxis(point.createdAt, 'hour')
                      }
                      return formatCalendarDayHeading(point.createdAt)
                    }}
                  />
                  <Bar
                    yAxisId="count"
                    dataKey="approvedCount"
                    fill={userChartColorApproved}
                    name="Approved requests"
                    radius={isXs ? [3, 3, 0, 0] : [4, 4, 0, 0]}
                    maxBarSize={userChartResponsive.maxBarSize}
                    minPointSize={0}
                  />
                  <Bar
                    yAxisId="amount"
                    dataKey="amount"
                    fill={userChartColorAmount}
                    name="Amount sum"
                    radius={isXs ? [3, 3, 0, 0] : [4, 4, 0, 0]}
                    maxBarSize={userChartResponsive.maxBarSize}
                    minPointSize={0}
                  />
                  <Line
                    yAxisId="count"
                    type="monotone"
                    dataKey="anomalyCount"
                    stroke={userChartColorAnomaly}
                    strokeWidth={isXs ? 1.5 : 2}
                    dot={{ r: ui.chartDotR, fill: userChartColorAnomaly }}
                    name="Anomaly count"
                  />
                </ComposedChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>
        <aside
          className={`flex min-w-0 shrink-0 border-base-300 text-base-content/90 lg:w-[7rem] lg:flex-col lg:justify-center lg:border-l lg:border-t-0 lg:pl-2 lg:pt-0 ${isDesktop ? '' : 'w-full flex-row flex-wrap items-center justify-center gap-x-4 gap-y-1 border-t pt-2'} ${isMd ? ui.tableCellText : 'text-[10px] leading-snug lg:text-xs'}`}
          aria-label="Chart series and axes"
        >
          <div className="flex flex-wrap items-center justify-center gap-x-3 gap-y-1 lg:flex-col lg:items-stretch lg:justify-center lg:gap-3 lg:space-y-0">
            <div className="flex min-w-0 flex-col gap-1 lg:gap-0">
              <p className="hidden font-semibold uppercase tracking-wide text-base-content/50 lg:mb-1 lg:block">
                Left axis
              </p>
              <div className="flex items-center gap-1.5">
                <span
                  className="h-3 w-3 shrink-0 rounded-sm"
                  style={{ backgroundColor: userChartColorAmount }}
                  aria-hidden
                />
                <span className="whitespace-nowrap" title="Daily amount sum (left axis)">
                  <span className="text-base-content/50 lg:hidden">L: </span>
                  Amount sum
                </span>
              </div>
            </div>
            <div className="flex min-w-0 flex-col gap-1 lg:gap-0">
              <p className="hidden font-semibold uppercase tracking-wide text-base-content/50 lg:mb-1 lg:block">
                Right axis
              </p>
              <div className="flex flex-wrap items-center justify-center gap-x-3 gap-y-1 lg:flex-col lg:items-stretch lg:gap-1.5">
                <div className="flex items-center gap-1.5">
                  <span
                    className="h-3 w-3 shrink-0 rounded-sm"
                    style={{ backgroundColor: userChartColorApproved }}
                    aria-hidden
                  />
                  <span className="whitespace-nowrap" title="Approved requests (right axis)">
                    <span className="text-base-content/50 lg:hidden">R: </span>
                    Approved
                  </span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="flex shrink-0 items-center gap-0.5" aria-hidden>
                    <span
                      className="h-0.5 w-2 rounded-full"
                      style={{ backgroundColor: userChartColorAnomaly }}
                    />
                    <span
                      className="h-1.5 w-1.5 shrink-0 rounded-full"
                      style={{ backgroundColor: userChartColorAnomaly }}
                    />
                  </span>
                  <span className="whitespace-nowrap" title="Anomaly count (right axis)">
                    <span className="text-base-content/50 lg:hidden">R: </span>
                    Anomalies
                  </span>
                </div>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </AppCard>
  )
}
