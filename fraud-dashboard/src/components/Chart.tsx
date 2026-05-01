import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { useMemo } from 'react'
import { useScreen } from '../config/ScreenContext'
import type { FraudRatePoint } from '../hooks/useLiveTransactions'

type ChartProps = {
  data: FraudRatePoint[]
  requestCount: number
  fraudCount: number
}

export default function Chart({ data, requestCount, fraudCount }: ChartProps) {
  const { isXs, isMd, isSm, ui } = useScreen()
  const xAxisTicks = useMemo(() => {
    if (data.length === 0) return []
    const ticks: string[] = []
    for (let i = 0; i < data.length; i += 60) {
      ticks.push(data[i].minuteLabel)
    }
    const lastTick = data[data.length - 1].minuteLabel
    if (ticks[ticks.length - 1] !== lastTick) {
      ticks.push(lastTick)
    }
    return ticks
  }, [data])

  return (
    <div className="card bg-base-100 shadow-md">
      <div className={`card-body ${ui.cardBodyPadding}`}>
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

        <div className={`${ui.chartHeight} w-full`}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart
              data={data}
              margin={{
                top: 8,
                right: isSm ? 4 : 8,
                bottom: isSm ? 0 : 8,
                left: isSm ? 0 : isMd ? -8 : 0,
              }}
            >
              {ui.chartGridVisible && <CartesianGrid strokeDasharray="3 3" />}
              <XAxis
                dataKey="minuteLabel"
                ticks={xAxisTicks}
                tickFormatter={(value: string) => value.slice(0, 5)}
                minTickGap={isSm ? 24 : isMd ? 20 : 32}
              />
              <YAxis
                unit="%"
                domain={[0, 100]}
                ticks={[0, 25, 50, 75, 100]}
                allowDecimals={false}
                width={isXs ? 30 : isSm ? 34 : 40}
              />
              <Tooltip
                filterNull={false}
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
      </div>
    </div>
  )
}
