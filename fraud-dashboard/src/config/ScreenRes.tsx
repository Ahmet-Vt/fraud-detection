export const ScreenRes = {
  xs: 480,
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
  '2xl': 1536,
} as const

export type ScreenResKey = keyof typeof ScreenRes

export const ScreenQuery = {
  xsUp: `(min-width: ${ScreenRes.xs}px)`,
  smUp: `(min-width: ${ScreenRes.sm}px)`,
  mdUp: `(min-width: ${ScreenRes.md}px)`,
  lgUp: `(min-width: ${ScreenRes.lg}px)`,
  xlUp: `(min-width: ${ScreenRes.xl}px)`,
  x2lUp: `(min-width: ${ScreenRes['2xl']}px)`,
  xsDown: `(max-width: ${ScreenRes.xs - 1}px)`,
  smDown: `(max-width: ${ScreenRes.sm - 1}px)`,
  mdDown: `(max-width: ${ScreenRes.md - 1}px)`,
  lgDown: `(max-width: ${ScreenRes.lg - 1}px)`,
  xlDown: `(max-width: ${ScreenRes.xl - 1}px)`,
} as const
