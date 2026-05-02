import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { ScreenRes } from './ScreenRes'

type ScreenState = {
  width: number
  isXs: boolean
  isSm: boolean
  isMd: boolean
  isLg: boolean
  isDesktop: boolean
  ui: {
    pagePadding: string
    containerWidth: string
    stackGap: string
    titleText: string
    cardBodyPadding: string
    cardTitleText: string
    badgeSize: string
    tableSize: string
    tableCellText: string
    chartHeight: string
    chartDotR: number
    chartActiveDotR: number
    chartGridVisible: boolean
  }
}

const defaultWidth = ScreenRes.lg

function getWidth() {
  if (typeof window === 'undefined') return defaultWidth
  return window.innerWidth
}

function toState(width: number): ScreenState {
  const isXs = width < ScreenRes.xs
  const isSm = width < ScreenRes.sm
  const isMd = width < ScreenRes.md
  const isLg = width < ScreenRes.lg
  const isDesktop = width >= ScreenRes.lg

  return {
    width,
    isXs,
    isSm,
    isMd,
    isLg,
    isDesktop,
    ui: {
      pagePadding: isXs ? 'p-2' : isSm ? 'p-3' : isLg ? 'p-6' : 'p-8',
      containerWidth: isMd ? 'max-w-full' : 'max-w-6xl',
      stackGap: isSm ? 'space-y-3' : 'space-y-4',
      titleText: isMd ? 'text-xl' : 'text-2xl',
      cardBodyPadding: isSm ? 'p-4' : 'p-6',
      cardTitleText: isMd ? 'text-base' : 'text-lg',
      badgeSize: isSm ? 'badge-sm' : 'badge-md',
      tableSize: isSm ? 'table-sm' : 'table-md',
      tableCellText: isSm ? 'text-xs' : 'text-sm',
      chartHeight: isSm ? 'h-56' : 'h-72',
      chartDotR: isMd ? 2 : 3,
      chartActiveDotR: isMd ? 4 : 5,
      chartGridVisible: !isMd,
    },
  }
}

const ScreenContext = createContext<ScreenState>(toState(defaultWidth))

export function ScreenProvider({ children }: { children: ReactNode }) {
  const [width, setWidth] = useState<number>(() => getWidth())

  useEffect(() => {
    const onResize = () => setWidth(window.innerWidth)
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  const value = useMemo(() => toState(width), [width])

  return <ScreenContext.Provider value={value}>{children}</ScreenContext.Provider>
}

export function useScreen() {
  return useContext(ScreenContext)
}
