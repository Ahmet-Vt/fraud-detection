import { useScreen } from '../config/ScreenContext'
import { formatLiveDateTime } from '../lib/formatDateTime'

export type LiveFraudData = {
  id: string
  userid: string
  amount: number
  location: string
  fraudstatus: boolean
  createdat: string
}

type LiveFraudDataProps = {
  data: LiveFraudData
}

export default function LiveFraudData({ data }: LiveFraudDataProps) {
  const { isMd, ui } = useScreen()
  const fraudLabel = data.fraudstatus ? 'anomaly' : 'approved'
  const fraudBadgeClass = data.fraudstatus
    ? 'badge badge-error badge-soft'
    : 'badge badge-success badge-soft'

  return (
    <div className="card bg-base-100 shadow-md">
      <div className={`card-body ${ui.cardBodyPadding}`}>
        <h2 className={`card-title ${ui.cardTitleText}`}>Live Fraud Data</h2>

        <div className={`grid grid-cols-1 gap-2 ${ui.tableCellText} ${isMd ? 'grid-cols-1' : 'grid-cols-2'}`}>
          <div>
            <span className="font-semibold">id:</span> {data.id}
          </div>
          <div>
            <span className="font-semibold">userid:</span> {data.userid}
          </div>
          <div>
            <span className="font-semibold">amount:</span> {data.amount}
          </div>
          <div>
            <span className="font-semibold">location:</span> {data.location}
          </div>
        </div>

        <div
          className={`mt-3 flex flex-col gap-2 border-t border-base-300 pt-3 sm:flex-row sm:items-center sm:justify-between ${ui.tableCellText}`}
        >
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <span className="font-semibold">fraudstatus:</span>
            <span className={fraudBadgeClass}>{fraudLabel}</span>
          </div>
          <div className="self-end shrink-0 whitespace-nowrap text-right sm:self-auto">
            <span className="font-semibold">createdat:</span> {formatLiveDateTime(data.createdat)}
          </div>
        </div>
      </div>
    </div>
  )
}
