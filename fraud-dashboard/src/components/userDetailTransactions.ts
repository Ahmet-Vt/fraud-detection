export type DateRangePreset = 'today' | 'weekly' | 'monthly' | 'yearly' | 'all'

export type UserTransaction = {
  id: string
  amount: number
  location: string
  fraudStatus: boolean
  createdAt: string
}

export type DailySummaryPoint = {
  bucketStart: string
  requestCount: number
  anomalyCount: number
  amountSum: number
  amountAvg: number
}

export function userDetailDateRangeQuery(preset: DateRangePreset): string {
  if (preset === 'all') return ''

  const now = new Date()
  let from: Date
  switch (preset) {
    case 'today':
      from = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0)
      break
    case 'weekly':
      from = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
      break
    case 'monthly':
      from = new Date(now.getFullYear(), now.getMonth(), 1, 0, 0, 0, 0)
      break
    case 'yearly':
      from = new Date(now.getFullYear(), 0, 1, 0, 0, 0, 0)
      break
    default:
      from = now
  }

  return new URLSearchParams({
    from: from.toISOString(),
    to: now.toISOString(),
  }).toString()
}
