import { memo, useCallback, useEffect, useId, useLayoutEffect, useRef, useState, type ReactNode } from 'react'
import notificationIcon from '../assets/notification.svg'

export type AnomalyInboxItem = {
  id: string
  userId: string
  userLabel: string
}

type AnomalyInboxFabProps = {
  items: AnomalyInboxItem[]
  icon?: ReactNode
  onRowClick?: (item: AnomalyInboxItem) => void
}

const rowText = 'text-[0.9375rem] leading-snug'
const rowPad = 'px-[7.5px] py-[5px]'

function AnomalyRowText({ userLabel }: { userLabel: string }) {
  return (
    <>
      Anomaly Request: There has been a suspicious transaction from{' '}
      <strong className="font-bold">{userLabel}</strong>
    </>
  )
}

function AnomalyInboxFab({ items, icon, onRowClick }: AnomalyInboxFabProps) {
  const [open, setOpen] = useState(false)
  const [pulseClass, setPulseClass] = useState(false)
  const rootRef = useRef<HTMLDivElement>(null)
  const listScrollRef = useRef<HTMLUListElement | null>(null)
  const prevListScrollRef = useRef<{ top: number; height: number }>({ top: 0, height: 0 })
  const prevTopIdRef = useRef<string | undefined>(undefined)
  const panelId = useId()

  useEffect(() => {
    const top = items[0]?.id
    if (top === undefined) return
    if (prevTopIdRef.current === top) return
    prevTopIdRef.current = top
    setPulseClass(false)
    requestAnimationFrame(() => {
      setPulseClass(true)
    })
  }, [items])

  useEffect(() => {
    if (!pulseClass) return
    const id = window.setTimeout(() => setPulseClass(false), 1400)
    return () => window.clearTimeout(id)
  }, [pulseClass])

  useLayoutEffect(() => {
    if (!open) {
      prevListScrollRef.current = { top: 0, height: 0 }
      return
    }
    const el = listScrollRef.current
    if (!el) return

    const prev = prevListScrollRef.current
    const nextHeight = el.scrollHeight
    const heightDelta = nextHeight - prev.height
    const wasNearTop = prev.top <= 8

    if (prev.height > 0 && heightDelta > 0 && !wasNearTop) {
      el.scrollTop = prev.top + heightDelta
    }

    prevListScrollRef.current = { top: el.scrollTop, height: el.scrollHeight }
  }, [items, open])

  const handleListScroll = useCallback(() => {
    const el = listScrollRef.current
    if (!el) return
    prevListScrollRef.current = {
      ...prevListScrollRef.current,
      top: el.scrollTop,
    }
  }, [])

  useEffect(() => {
    if (!open) return
    const onPointerDown = (e: MouseEvent | TouchEvent) => {
      const el = rootRef.current
      if (!el) return
      const target = e.target as Node
      if (!el.contains(target)) setOpen(false)
    }
    document.addEventListener('mousedown', onPointerDown)
    document.addEventListener('touchstart', onPointerDown)
    return () => {
      document.removeEventListener('mousedown', onPointerDown)
      document.removeEventListener('touchstart', onPointerDown)
    }
  }, [open])

  const count = items.length

  return (
    <div ref={rootRef} className="pointer-events-auto fixed bottom-6 end-6 z-50 flex flex-col items-end gap-2">
      {open ? (
        <div
          id={panelId}
          role="dialog"
          aria-label="Anomaly notifications"
          className="card mb-1 flex max-h-[min(calc(30rem),calc(50vh))] min-h-0 w-[min(22rem,calc(100vw-2rem))] flex-col overflow-hidden border border-base-300 bg-base-100 shadow-xl"
        >
          <div className="card-body flex min-h-0 flex-1 flex-col gap-1.5 overflow-hidden p-2.5 pt-2">
            <h2 className="card-title shrink-0 text-[1.09375rem] font-semibold leading-tight">Anomaly alerts</h2>
            <ul
              ref={listScrollRef}
              onScroll={handleListScroll}
              className="flex min-h-0 flex-1 flex-col gap-px overflow-y-auto rounded-md bg-base-300/40 p-px [overflow-anchor:none]"
            >
              {items.length === 0 ? (
                <li className={`rounded-sm ${rowPad} ${rowText} opacity-70`}>No anomalies yet.</li>
              ) : (
                items.map((item) => (
                  <li key={item.id} className="w-full">
                    {onRowClick ? (
                      <button
                        type="button"
                        className={`w-full whitespace-normal rounded-sm bg-red-600 ${rowPad} text-left ${rowText} text-white hover:bg-red-700`}
                        onMouseDown={(e) => {
                          if (e.button !== 0) return
                          e.preventDefault()
                          onRowClick(item)
                          setOpen(false)
                        }}
                        onKeyDown={(e) => {
                          if (e.key !== 'Enter' && e.key !== ' ') return
                          e.preventDefault()
                          onRowClick(item)
                          setOpen(false)
                        }}
                      >
                        <AnomalyRowText userLabel={item.userLabel} />
                      </button>
                    ) : (
                      <span
                        className={`block rounded-sm bg-red-600 ${rowPad} ${rowText} text-white`}
                      >
                        <AnomalyRowText userLabel={item.userLabel} />
                      </span>
                    )}
                  </li>
                ))
              )}
            </ul>
          </div>
        </div>
      ) : null}

      <button
        type="button"
        className={`btn btn-circle relative h-14 w-14 shrink-0 border-2 border-amber-500/80 bg-amber-300 text-amber-950 shadow-lg hover:bg-amber-400 ${
          pulseClass ? 'anomaly-bell-pulse' : ''
        }`}
        aria-label="Anomaly notifications"
        aria-expanded={open}
        aria-controls={open ? panelId : undefined}
        aria-haspopup="dialog"
        onClick={() => setOpen((v) => !v)}
      >
        {count > 0 ? (
          <span className="badge badge-error badge-sm absolute -end-1 -top-1 min-w-[1.25rem] border-2 border-base-100 px-1 font-mono text-[10px]">
            {count > 99 ? '99+' : count}
          </span>
        ) : null}
        <span className="flex h-8 w-8 items-center justify-center [&>img]:h-8 [&>img]:w-8 [&>svg]:h-8 [&>svg]:w-8">
          {icon ?? <img src={notificationIcon} alt="" className="h-8 w-8 object-contain" />}
        </span>
      </button>
    </div>
  )
}

export default memo(AnomalyInboxFab)
