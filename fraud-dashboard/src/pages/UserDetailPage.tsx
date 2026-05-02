import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { apiUserMemberUrl } from '../api/endpoints'
import UserTransactionsPanel from '../components/UserTransactionsPanel'
import type { UserRow } from '../components/UsersTable'

export default function UserDetailPage() {
  const { userId } = useParams<{ userId: string }>()

  const [user, setUser] = useState<UserRow | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!userId) {
      setIsLoading(false)
      setError('Missing user id')
      return
    }

    let cancelled = false
    const controller = new AbortController()

    const loadUser = async () => {
      setIsLoading(true)
      setError(null)
      try {
        const response = await fetch(apiUserMemberUrl(userId), {
          signal: controller.signal,
        })
        if (!response.ok) {
          if (response.status === 404) {
            throw new Error('User not found')
          }
          throw new Error(`User request failed with ${response.status}`)
        }

        const data = (await response.json()) as {
          id: string
          name: string
          surname: string
          email?: string
          status: string
          created_at: string
        }

        if (cancelled) return

        setUser({
          id: data.id,
          name: data.name,
          surname: data.surname,
          email: data.email,
          status: data.status,
          createdAt: data.created_at,
        })
      } catch (err) {
        if (cancelled) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setUser(null)
        setError(err instanceof Error ? err.message : 'Could not load user')
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }

    void loadUser()

    return () => {
      cancelled = true
      controller.abort()
    }
  }, [userId])

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
