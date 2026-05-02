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

/** Inclusive calendar-day range in the user's local timezone, or all-time. */
export type UserTransactionDateRange =
  | { kind: 'all' }
  | { kind: 'between'; fromDay: string; toDay: string }

const dayKeyRe = /^\d{4}-\d{2}-\d{2}$/

export function formatLocalDayKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export function parseLocalDay(dayKey: string): Date | null {
  if (!dayKeyRe.test(dayKey)) return null
  const [y, mo, d] = dayKey.split('-').map(Number)
  const dt = new Date(y, mo - 1, d)
  if (dt.getFullYear() !== y || dt.getMonth() !== mo - 1 || dt.getDate() !== d) return null
  return dt
}

export function startOfLocalDayFromKey(dayKey: string): Date {
  const d = parseLocalDay(dayKey)
  if (!d) return new Date(0)
  return new Date(d.getFullYear(), d.getMonth(), d.getDate(), 0, 0, 0, 0)
}

export function endOfLocalDayFromKey(dayKey: string): Date {
  const d = parseLocalDay(dayKey)
  if (!d) return new Date(0)
  return new Date(d.getFullYear(), d.getMonth(), d.getDate(), 23, 59, 59, 999)
}

function pickBucket(from: Date, to: Date): string {
  const hours = (to.getTime() - from.getTime()) / (3600 * 1000)
  return hours <= 48 ? 'hour' : 'day'
}

/** Query string for `.../detail` (includes bucket). */
export function userTransactionDetailQuery(range: UserTransactionDateRange): string {
  if (range.kind === 'all') {
    return 'bucket=day'
  }
  const from = startOfLocalDayFromKey(range.fromDay)
  const to = endOfLocalDayFromKey(range.toDay)
  const bucket = pickBucket(from, to)
  return new URLSearchParams({
    from: from.toISOString(),
    to: to.toISOString(),
    bucket,
  }).toString()
}

export function rangeAll(): UserTransactionDateRange {
  return { kind: 'all' }
}

export function rangeLastNDays(n: number): UserTransactionDateRange {
  const to = new Date()
  const from = new Date(to)
  from.setDate(from.getDate() - (n - 1))
  return { kind: 'between', fromDay: formatLocalDayKey(from), toDay: formatLocalDayKey(to) }
}

export function rangeToday(): UserTransactionDateRange {
  const t = new Date()
  const day = formatLocalDayKey(t)
  return { kind: 'between', fromDay: day, toDay: day }
}

export function rangeThisMonth(): UserTransactionDateRange {
  const now = new Date()
  const from = new Date(now.getFullYear(), now.getMonth(), 1)
  return { kind: 'between', fromDay: formatLocalDayKey(from), toDay: formatLocalDayKey(now) }
}

export function rangeThisYear(): UserTransactionDateRange {
  const now = new Date()
  const from = new Date(now.getFullYear(), 0, 1)
  return { kind: 'between', fromDay: formatLocalDayKey(from), toDay: formatLocalDayKey(now) }
}

export function isValidBetweenRange(fromDay: string, toDay: string): boolean {
  if (!dayKeyRe.test(fromDay) || !dayKeyRe.test(toDay)) return false
  const a = parseLocalDay(fromDay)
  const b = parseLocalDay(toDay)
  if (!a || !b) return false
  return a.getTime() <= b.getTime()
}
