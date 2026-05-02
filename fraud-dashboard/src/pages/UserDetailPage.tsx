import { useParams } from 'react-router-dom'
import UserTransactionsPanel from '../components/UserTransactionsPanel'
import { useUserMember } from '../hooks/useUserMember'

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

  return <UserTransactionsPanel user={user} />
}
