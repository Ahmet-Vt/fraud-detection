import { useCallback, useEffect, useRef, useState, type SVGProps } from 'react'
import type { DateRange } from 'react-day-picker'
import { DayPicker } from 'react-day-picker'
import {
  formatLocalDayKey,
  parseLocalDay,
  rangeAll,
  rangeLastNDays,
  rangeThisMonth,
  rangeThisYear,
  rangeToday,
  type UserTransactionDateRange,
} from './userDetailTransactions'

function startOfCurrentMonth(): Date {
  const n = new Date()
  return new Date(n.getFullYear(), n.getMonth(), 1)
}

function CalendarGlyph(props: SVGProps<SVGSVGElement>) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2} aria-hidden {...props}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
    </svg>
  )
}

function CloseGlyph(props: SVGProps<SVGSVGElement>) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2} aria-hidden {...props}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
    </svg>
  )
}

function formatDateRangeLabel(range: UserTransactionDateRange): string {
  if (range.kind === 'all') return 'All time'
  const a = parseLocalDay(range.fromDay)
  const b = parseLocalDay(range.toDay)
  if (!a || !b) return 'Date range'
  if (range.fromDay === range.toDay) {
    return a.toLocaleDateString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }
  return `${a.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })} – ${b.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}`
}

type UserTransactionRangePickerButtonProps = {
  dateRange: UserTransactionDateRange
  onDateRangeChange: (range: UserTransactionDateRange) => void
  isLoading: boolean
  buttonSizeClass: string
  textClass: string
}

export default function UserTransactionRangePickerButton({
  dateRange,
  onDateRangeChange,
  isLoading,
  buttonSizeClass,
  textClass,
}: UserTransactionRangePickerButtonProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const [pickerRange, setPickerRange] = useState<DateRange | undefined>()
  const [applyError, setApplyError] = useState<string | null>(null)
  const [calendarMonth, setCalendarMonth] = useState(() => startOfCurrentMonth())

  const syncPickerFromCommitted = useCallback(() => {
    if (dateRange.kind === 'all') {
      setPickerRange(undefined)
    } else {
      const from = parseLocalDay(dateRange.fromDay)
      const to = parseLocalDay(dateRange.toDay)
      if (from && to) setPickerRange({ from, to })
    }
    setApplyError(null)
  }, [dateRange])

  const openDialog = () => {
    syncPickerFromCommitted()
    setCalendarMonth(startOfCurrentMonth())
    dialogRef.current?.showModal()
  }

  useEffect(() => {
    const el = dialogRef.current
    if (!el) return
    const onClose = () => {
      syncPickerFromCommitted()
    }
    el.addEventListener('close', onClose)
    return () => el.removeEventListener('close', onClose)
  }, [syncPickerFromCommitted])

  const closeDialog = () => {
    dialogRef.current?.close()
  }

  const applyCalendarRange = () => {
    if (!pickerRange?.from || !pickerRange?.to) {
      setApplyError('Select a start and end date on the calendar.')
      return
    }
    setApplyError(null)
    onDateRangeChange({
      kind: 'between',
      fromDay: formatLocalDayKey(pickerRange.from),
      toDay: formatLocalDayKey(pickerRange.to),
    })
    closeDialog()
  }

  const applyPreset = (range: UserTransactionDateRange) => {
    setApplyError(null)
    onDateRangeChange(range)
    closeDialog()
  }

  const chipClass = `btn btn-outline btn-xs shrink-0 sm:btn-sm ${textClass}`

  return (
    <>
      <button
        type="button"
        className={`btn btn-primary w-full min-h-11 justify-between gap-2 normal-case sm:w-auto sm:min-w-[12rem] ${buttonSizeClass}`}
        disabled={isLoading}
        onClick={openDialog}
        aria-haspopup="dialog"
      >
        <span className={`truncate text-left ${textClass}`}>{formatDateRangeLabel(dateRange)}</span>
        <CalendarGlyph className="h-5 w-5 shrink-0 opacity-90" />
      </button>

      <dialog ref={dialogRef} className="modal modal-bottom sm:modal-middle" onCancel={() => closeDialog()}>
        <div className="modal-box box-border max-h-[min(90dvh,32rem)] w-[min(100vw-1rem,24rem)] max-w-none overflow-y-auto overscroll-contain p-4 sm:max-h-[85vh] sm:w-[min(100vw-2rem,26rem)] sm:p-5">
          <header className="mb-3 flex items-start justify-between gap-2">
            <h4 className={`font-semibold leading-snug ${textClass}`}>Choose date range</h4>
            <button type="button" className="btn btn-ghost btn-sm btn-square shrink-0" aria-label="Close" onClick={closeDialog}>
              <CloseGlyph className="h-5 w-5" />
            </button>
          </header>

          <div className="flex min-w-0 flex-wrap gap-1.5 pb-3 sm:gap-2" role="group" aria-label="Quick ranges">
            <button type="button" className={chipClass} onClick={() => applyPreset(rangeToday())}>
              Today
            </button>
            <button type="button" className={chipClass} onClick={() => applyPreset(rangeLastNDays(7))}>
              Last 7 days
            </button>
            <button type="button" className={chipClass} onClick={() => applyPreset(rangeThisMonth())}>
              This month
            </button>
            <button type="button" className={chipClass} onClick={() => applyPreset(rangeThisYear())}>
              This year
            </button>
            <button type="button" className={chipClass} onClick={() => applyPreset(rangeAll())}>
              All time
            </button>
          </div>

          <div className="flex min-w-0 justify-center overflow-x-auto rounded-lg border border-base-300 bg-base-100 p-1 sm:p-2">
            <DayPicker
              mode="range"
              month={calendarMonth}
              onMonthChange={setCalendarMonth}
              selected={pickerRange}
              onSelect={(next) => {
                setApplyError(null)
                setPickerRange(next)
              }}
              numberOfMonths={1}
              showOutsideDays
              className="mx-auto [--rdp-accent-color:oklch(var(--p))] [--rdp-background-color:oklch(var(--b1))]"
            />
          </div>

          {applyError && (
            <p className={`mt-2 text-error ${textClass}`} role="alert">
              {applyError}
            </p>
          )}

          <div className="modal-action mt-4 flex-wrap gap-2 sm:mt-5">
            <button type="button" className="btn btn-ghost btn-sm sm:btn-md" onClick={closeDialog}>
              Cancel
            </button>
            <button type="button" className="btn btn-primary btn-sm sm:btn-md" onClick={applyCalendarRange}>
              Apply range
            </button>
          </div>
        </div>
        <form method="dialog" className="modal-backdrop">
          <button type="submit" aria-label="Close date picker">
            close
          </button>
        </form>
      </dialog>
    </>
  )
}
