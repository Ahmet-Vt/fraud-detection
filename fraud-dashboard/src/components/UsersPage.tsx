import { useState } from 'react'
import { useScreen } from '../config/ScreenContext'
import UserTransactionsPanel from './UserTransactionsPanel'
import UsersTable, { type UserRow } from './UsersTable'

type UsersPageProps = {
  usersUrl: string
  transactionsBaseUrl?: string
}

export default function UsersPage({ usersUrl, transactionsBaseUrl = '/transactions/user' }: UsersPageProps) {
  const { ui } = useScreen()
  const [selectedUser, setSelectedUser] = useState<UserRow | null>(null)

  if (selectedUser) {
    return (
      <UserTransactionsPanel
        user={selectedUser}
        onBack={() => setSelectedUser(null)}
        transactionsBaseUrl={transactionsBaseUrl}
      />
    )
  }

  return (
    <div className={`space-y-3 ${ui.stackGap}`}>
      <UsersTable usersUrl={usersUrl} onSelectUser={setSelectedUser} />
    </div>
  )
}
