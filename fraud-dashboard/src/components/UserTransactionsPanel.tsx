import { useMemo, useState } from 'react'
import { useScreen } from '../config/ScreenContext'
import { useUserTransactionDetail } from '../hooks/useUserTransactionDetail'
import type { DateRangePreset } from './userDetailTransactions'
import type { UserRow } from './UsersTable'
import UserTransactionsChartCard from './UserTransactionsChartCard'
import UserTransactionsTableCard from './UserTransactionsTableCard'

type UserTransactionsPanelProps = {
  user: UserRow
}

export default function UserTransactionsPanel({ user }: UserTransactionsPanelProps) {
  const { ui } = useScreen()
  const [datePreset, setDatePreset] = useState<DateRangePreset>('all')
  const { transactions, dailySummary, isLoading, error } = useUserTransactionDetail(user.id, datePreset)

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
        anomalyRatio={anomalyRatio}
        isLoading={isLoading}
        error={error}
      />
      <UserTransactionsTableCard
        transactions={transactions}
        datePreset={datePreset}
        onDatePresetChange={setDatePreset}
        isLoading={isLoading}
      />
    </div>
  )
}
