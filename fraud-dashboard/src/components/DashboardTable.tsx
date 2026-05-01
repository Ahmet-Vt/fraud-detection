import { useScreen } from '../config/ScreenContext'

type TransactionRow = {
  id: string
  userId: string
  amount: number
  location: string
  status: 'completed' | 'failed' | 'pending'
  fraud: boolean
}

const rows: TransactionRow[] = [
  {
    id: 'tx-001',
    userId: 'u-1001',
    amount: 4200,
    location: 'Istanbul',
    status: 'completed',
    fraud: false,
  },
  {
    id: 'tx-002',
    userId: 'u-1001',
    amount: 46000,
    location: 'Van',
    status: 'completed',
    fraud: true,
  },
  {
    id: 'tx-003',
    userId: 'u-2099',
    amount: 980,
    location: 'Ankara',
    status: 'pending',
    fraud: false,
  },
  {
    id: 'tx-004',
    userId: 'u-3340',
    amount: 12500,
    location: 'Izmir',
    status: 'failed',
    fraud: true,
  },
]

function statusClass(status: TransactionRow['status']) {
  if (status === 'completed') return 'badge badge-success badge-soft'
  if (status === 'failed') return 'badge badge-error badge-soft'
  return 'badge badge-warning badge-soft'
}

export default function DashboardTable() {
  const { isMd, ui } = useScreen()

  return (
    <div className="card bg-base-100 shadow-md">
      <div className={`card-body ${ui.cardBodyPadding}`}>
        <div className="flex items-center justify-between gap-2">
          <h2 className={`card-title ${ui.cardTitleText}`}>Live Transaction Feed</h2>
          <span className={`badge badge-info badge-outline ${ui.badgeSize}`}>Last 4 events</span>
        </div>

        <div className="overflow-x-auto">
          <table className={`table table-zebra ${ui.tableSize}`}>
            <thead>
              <tr className={ui.tableCellText}>
                <th>Transaction</th>
                <th>User</th>
                <th>Amount</th>
                {!isMd && <th>Location</th>}
                <th>Status</th>
                <th>Fraud</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.id} className={ui.tableCellText}>
                  <td className="font-medium">{row.id}</td>
                  <td>{row.userId}</td>
                  <td>${row.amount.toLocaleString()}</td>
                  {!isMd && <td>{row.location}</td>}
                  <td>
                    <span className={`${statusClass(row.status)} ${ui.badgeSize}`}>{row.status}</span>
                  </td>
                  <td>
                    <span
                      className={`${row.fraud ? 'badge badge-error' : 'badge badge-success'} ${ui.badgeSize}`}
                    >
                      {row.fraud ? 'suspicious' : 'approved'}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
