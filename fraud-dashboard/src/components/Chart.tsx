import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'

type DailyMetric = {
  day: string
  suspiciousRate: number
}

const weeklyData: DailyMetric[] = [
  { day: 'Mon', suspiciousRate: 8.2 },
  { day: 'Tue', suspiciousRate: 6.9 },
  { day: 'Wed', suspiciousRate: 9.4 },
  { day: 'Thu', suspiciousRate: 7.8 },
  { day: 'Fri', suspiciousRate: 11.1 },
  { day: 'Sat', suspiciousRate: 10.2 },
  { day: 'Sun', suspiciousRate: 8.7 },
]

export default function Chart() {
  return (
    <div className="card bg-base-100 shadow-md">
      <div className="card-body">
        <div className="flex items-center justify-between gap-2">
          <h2 className="card-title">7-Day Suspicious Rate</h2>
          <span className="badge badge-outline">Daily</span>
        </div>

        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart
              data={weeklyData}
              margin={{ top: 8, right: 8, bottom: 8, left: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="day" />
              <YAxis unit="%" domain={[0, 15]} />
              <Tooltip
                formatter={(value) => [`${value ?? 0}%`, 'Suspicious Rate']}
              />
              <Line
                type="monotone"
                dataKey="suspiciousRate"
                stroke="#ef4444"
                strokeWidth={2}
                dot={{ r: 3 }}
                activeDot={{ r: 5 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}
