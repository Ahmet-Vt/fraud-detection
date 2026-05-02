import type { ReactNode } from 'react'

export function tableScrollAreaClass(isMd: boolean, options?: { overflowAnchorNone?: boolean }): string {
  const base = `min-w-0 max-w-full ${isMd ? 'max-h-[22rem]' : 'max-h-[28rem]'} overflow-x-auto overflow-y-auto`
  return options?.overflowAnchorNone ? `${base} [overflow-anchor:none]` : base
}

type AppCardProps = {
  children: ReactNode
  bodyPaddingClass: string
  className?: string
  bodyClassName?: string
}

export function AppCard({ children, bodyPaddingClass, className, bodyClassName }: AppCardProps) {
  return (
    <div className={['card bg-base-100 shadow-md', className].filter(Boolean).join(' ')}>
      <div className={['card-body', bodyPaddingClass, bodyClassName].filter(Boolean).join(' ')}>{children}</div>
    </div>
  )
}

type InlineAlertProps = {
  variant: 'error' | 'info'
  textClass: string
  children: ReactNode
  className?: string
}

export function InlineAlert({ variant, textClass, children, className }: InlineAlertProps) {
  const variantClass = variant === 'error' ? 'alert alert-error' : 'alert alert-info'
  return <div className={[variantClass, textClass, className].filter(Boolean).join(' ')}>{children}</div>
}

type SimplePaginationProps = {
  page: number
  totalPages: number
  isMd: boolean
  textClass: string
  pageSize?: number
  onPrev: () => void
  onNext: () => void
}

export function SimplePagination({
  page,
  totalPages,
  isMd,
  textClass,
  pageSize = 100,
  onPrev,
  onNext,
}: SimplePaginationProps) {
  const btnSize = isMd ? 'btn-xs' : 'btn-sm'
  return (
    <div className="mt-3 flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
      <span className={`min-w-0 ${textClass}`}>
        Page {page} / {totalPages} - {pageSize} rows per page
      </span>
      <div className="join shrink-0">
        <button type="button" className={`btn join-item ${btnSize}`} disabled={page === 1} onClick={onPrev}>
          Prev
        </button>
        <button
          type="button"
          className={`btn join-item ${btnSize}`}
          disabled={page >= totalPages}
          onClick={onNext}
        >
          Next
        </button>
      </div>
    </div>
  )
}

export type FraudStatusVariant = 'liveTable' | 'livePanel' | 'userDetailPill'

type FraudStatusBadgeProps = {
  isFraud: boolean
  variant: FraudStatusVariant
  /** Required for `liveTable` — e.g. `ui.badgeSize` */
  badgeSizeClass?: string
}

export function FraudStatusBadge({ isFraud, variant, badgeSizeClass }: FraudStatusBadgeProps) {
  if (variant === 'userDetailPill') {
    return (
      <span
        className={`inline-flex h-6 w-[5.75rem] shrink-0 items-center justify-center rounded-[8px] text-xs font-medium text-white ${
          isFraud ? 'bg-red-600' : 'bg-green-600'
        }`}
      >
        {isFraud ? 'Anomaly' : 'Approved'}
      </span>
    )
  }

  if (variant === 'livePanel') {
    const fraudBadgeClass = isFraud ? 'badge badge-error badge-soft' : 'badge badge-success badge-soft'
    return <span className={fraudBadgeClass}>{isFraud ? 'anomaly' : 'approved'}</span>
  }

  const tableClass = isFraud
    ? 'badge bg-red-600 text-white border-red-700'
    : 'badge badge-success badge-soft'
  return (
    <span className={[tableClass, badgeSizeClass].filter(Boolean).join(' ')}>
      {isFraud ? 'anomaly' : 'approved'}
    </span>
  )
}
