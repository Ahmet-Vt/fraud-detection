export function formatUserRiskRatePercent(rate: number | undefined, maxFractionDigits = 2): string {
  const n = rate == null || !Number.isFinite(rate) ? 0 : rate
  const clamped = Math.min(100, Math.max(0, n))
  const formatted = new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: maxFractionDigits,
  }).format(clamped)
  return `%${formatted}`
}
