import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScreen } from '../config/ScreenContext'

type UsersTableProps = {
  usersUrl: string
}

export type UserRow = {
  id: string
  name: string
  surname: string
  email?: string
  status: string
  createdAt: string
}

export default function UsersTable({ usersUrl }: UsersTableProps) {
  const navigate = useNavigate()
  const { isMd, ui } = useScreen()
  const [users, setUsers] = useState<UserRow[]>([])
  const [isUsersLoading, setIsUsersLoading] = useState(false)
  const [usersError, setUsersError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const pageSize = 100

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
          payload.map((user) => ({
            id: user.id,
            name: user.name,
            surname: user.surname,
            email: user.email,
            status: user.status,
            createdAt: user.created_at,
          })),
        )
        setPage(1)
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

  const getStatusBadgeClass = (status: string) => {
    const normalized = status.trim().toLowerCase()
    if (normalized === 'active') return 'badge badge-success badge-soft'
    if (normalized === 'suspended') return 'badge badge-warning badge-soft'
    if (normalized === 'deleted') return 'badge badge-error badge-soft'
    return 'badge badge-ghost'
  }

  const formatDate = (value: string) => {
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return '-'
    const dd = String(date.getDate()).padStart(2, '0')
    const mm = String(date.getMonth() + 1).padStart(2, '0')
    const yyyy = String(date.getFullYear())
    return `${dd}:${mm}:${yyyy}`
  }

  const totalPages = Math.max(1, Math.ceil(users.length / pageSize))
  const pagedUsers = users.slice((page - 1) * pageSize, page * pageSize)

  return (
    <div className="card bg-base-100 shadow-md">
      <div className={`card-body ${ui.cardBodyPadding}`}>
        <div className="flex items-center justify-between gap-2">
          <h2 className={`card-title ${ui.cardTitleText}`}>Users</h2>
          <span className={`badge badge-outline ${ui.badgeSize}`}>{users.length} users</span>
        </div>
        {usersError && <div className={`alert alert-error ${ui.tableCellText}`}>{usersError}</div>}
        {isUsersLoading && <div className={`alert alert-info ${ui.tableCellText}`}>Loading users...</div>}
        <div className={`${isMd ? 'max-h-[22rem]' : 'max-h-[28rem]'} overflow-auto`}>
          <table className={`table table-zebra ${ui.tableSize}`}>
            <thead>
              <tr className={`sticky top-0 z-10 bg-base-100 ${ui.tableCellText}`}>
                <th>User</th>
                <th>User ID</th>
                {!isMd && <th>Email</th>}
                <th>Status</th>
                <th>Created At</th>
              </tr>
            </thead>
            <tbody>
              {!isUsersLoading && users.length === 0 ? (
                <tr>
                  <td colSpan={isMd ? 4 : 5} className="text-center opacity-70">
                    No users found.
                  </td>
                </tr>
              ) : (
                pagedUsers.map((row) => (
                  <tr
                    key={row.id}
                    className={`${ui.tableCellText} cursor-pointer hover:bg-base-200`}
                    onClick={() => navigate(`/users/${encodeURIComponent(row.id)}`)}
                  >
                    <td>{`${row.name} ${row.surname}`.trim()}</td>
                    <td className="font-medium">{row.id}</td>
                    {!isMd && <td>{row.email ?? '-'}</td>}
                    <td>
                      <span className={`${getStatusBadgeClass(row.status)} ${ui.badgeSize}`}>{row.status}</span>
                    </td>
                    <td>{formatDate(row.createdAt)}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {users.length > 0 && (
          <div className="mt-3 flex items-center justify-between">
            <span className={ui.tableCellText}>
              Page {page} / {totalPages} - 100 rows per page
            </span>
            <div className="join">
              <button
                type="button"
                className={`btn join-item ${isMd ? 'btn-xs' : 'btn-sm'}`}
                disabled={page === 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                Prev
              </button>
              <button
                type="button"
                className={`btn join-item ${isMd ? 'btn-xs' : 'btn-sm'}`}
                disabled={page >= totalPages}
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
