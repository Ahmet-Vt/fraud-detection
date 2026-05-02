function parseDate(value: string): Date | null {
  if (!value?.trim()) return null
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? null : d
}

function pad2(n: number): string {
  return String(n).padStart(2, '0')
}

export function formatDateTime(value: string): string {
  const d = parseDate(value)
  if (!d) return '—'
  return `${pad2(d.getDate())}.${pad2(d.getMonth() + 1)}.${d.getFullYear()}`
}

export function formatLiveDateTime(value: string): string {
  const d = parseDate(value)
  if (!d) return '—'
  return `${pad2(d.getDate())}.${pad2(d.getMonth() + 1)} ${pad2(d.getHours())}:${pad2(d.getMinutes())}:${pad2(d.getSeconds())}`
}

export function localCalendarDayKey(value: string): string | null {
  const d = parseDate(value)
  if (!d) return null
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`
}

export function formatCalendarDayHeading(value: string): string {
  const d = parseDate(value)
  if (!d) return '—'
  return d.toLocaleDateString(undefined, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

export function formatChartBucketAxis(value: string, bucket: 'day' | 'hour'): string {
  const d = parseDate(value)
  if (!d) return '—'
  if (bucket === 'day') {
    return `${pad2(d.getDate())}.${pad2(d.getMonth() + 1)}.${d.getFullYear()}`
  }
  return `${pad2(d.getDate())}.${pad2(d.getMonth() + 1)} ${pad2(d.getHours())}:${pad2(d.getMinutes())}`
}
