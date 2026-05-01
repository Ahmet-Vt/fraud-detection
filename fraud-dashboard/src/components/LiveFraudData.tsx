export type LiveFraudData = {
  id: string
  userid: string
  amount: number
  location: string
  transactionStatus: string
  fraudstatus: boolean
  createdat: string
}

type LiveFraudDataProps = {
  data: LiveFraudData
}

export default function LiveFraudData({ data }: LiveFraudDataProps) {
  const fraudBadgeClass = data.fraudstatus
    ? 'badge badge-error badge-soft'
    : 'badge badge-success badge-soft'

  return (
    <div className="card bg-base-100 shadow-md">
      <div className="card-body p-4 md:p-6">
        <h2 className="card-title text-base md:text-lg">Live Fraud Data</h2>

        <div className="grid grid-cols-1 gap-2 text-sm md:grid-cols-2">
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
          <div>
            <span className="font-semibold">transactionStatus:</span> {data.transactionStatus}
          </div>
          <div className="flex items-center gap-2">
            <span className="font-semibold">fraudstatus:</span>
            <span className={fraudBadgeClass}>{String(data.fraudstatus)}</span>
          </div>
          <div className="md:col-span-2">
            <span className="font-semibold">createdat:</span> {data.createdat}
          </div>
        </div>
      </div>
    </div>
  )
}
