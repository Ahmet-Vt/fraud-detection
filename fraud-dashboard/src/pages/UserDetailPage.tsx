import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useScreen } from '../config/ScreenContext'
import UserTransactionsChartCard from '../features/users/UserTransactionsChartCard'
import UserTransactionsTableCard from '../features/users/UserTransactionsTableCard'
import { rangeAll, type UserTransactionDateRange } from '../features/users/userDetailTransactions'
import type { UserRow } from '../features/users/userTypes'
import { useUserMember } from '../features/users/hooks/useUserMember'
import { useUserTransactionDetail } from '../features/users/hooks/useUserTransactionDetail'

function UserDetailTransactionsBody({ user }: { user: UserRow }) {
  const { ui } = useScreen()
  const [dateRange, setDateRange] = useState<UserTransactionDateRange>(() => rangeAll())
  const { transactions, dailySummary, isLoading, error } = useUserTransactionDetail(user.id, dateRange)

  const anomalyCount = useMemo(
    () => transactions.reduce((acc, tx) => acc + (tx.fraudStatus ? 1 : 0), 0),
    [transactions],
  )
  const anomalyRatio = transactions.length === 0 ? 0 : (anomalyCount / transactions.length) * 100

  return (
    <div className={`space-y-3 ${ui.stackGap}`}>
      <UserTransactionsChartCard
        user={user}
        dailySummary={dailySummary}
        dateRange={dateRange}
        anomalyRatio={anomalyRatio}
        isLoading={isLoading}
        error={error}
      />
      <UserTransactionsTableCard
        transactions={transactions}
        dateRange={dateRange}
        onDateRangeChange={setDateRange}
        isLoading={isLoading}
      />
    </div>
  )
}

export default function UserDetailPage() {
  const { userId } = useParams<{ userId: string }>()
  const { user, isLoading, error } = useUserMember(userId)

  if (!userId) {
    return null
  }

  if (isLoading) {
    return <div className="alert alert-info">Loading user...</div>
  }

  if (error || !user) {
    return (
      <div className="space-y-3">
        <div className="alert alert-error">{error ?? 'User not found'}</div>
      </div>
    )
  }

  return <UserDetailTransactionsBody user={user} />
}
