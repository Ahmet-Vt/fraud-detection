import { useEffect, useState } from 'react'
import type { UserRow } from '../components/UsersTable'

export function useUsersList(usersUrl: string) {
  const [users, setUsers] = useState<UserRow[]>([])
  const [isUsersLoading, setIsUsersLoading] = useState(false)
  const [usersError, setUsersError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    const controller = new AbortController()

    const loadUsers = async () => {
      setIsUsersLoading(true)
      setUsersError(null)
      try {
        const response = await fetch(usersUrl, { signal: controller.signal })
        if (!response.ok) {
          throw new Error(`users request failed with ${response.status}`)
        }

        const payload = (await response.json()) as Array<{
          id: string
          name: string
          surname: string
          email?: string
          status: string
          created_at: string
        }>

        if (cancelled) return

        setUsers(
          payload.map((u) => ({
            id: u.id,
            name: u.name,
            surname: u.surname,
            email: u.email,
            status: u.status,
            createdAt: u.created_at,
          })),
        )
      } catch (err) {
        if (cancelled) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setUsersError('Could not load users')
      } finally {
        if (!cancelled) {
          setIsUsersLoading(false)
        }
      }
    }

    void loadUsers()

    return () => {
      cancelled = true
      controller.abort()
    }
  }, [usersUrl])

  return { users, isUsersLoading, usersError }
}
