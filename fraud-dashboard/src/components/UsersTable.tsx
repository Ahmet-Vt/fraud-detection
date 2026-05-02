import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScreen } from '../config/ScreenContext'
import { useUsersList } from '../hooks/useUsersList'
import { formatDateTime } from '../lib/formatDateTime'

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
  const { users, isUsersLoading, usersError } = useUsersList(usersUrl)
  const [page, setPage] = useState(1)
  const pageSize = 100

  useEffect(() => {
    setPage(1)
  }, [users])

  const getStatusBadgeClass = (status: string) => {
    const normalized = status.trim().toLowerCase()
    if (normalized === 'active') return 'badge badge-success badge-soft'
    if (normalized === 'suspended') return 'badge badge-warning badge-soft'
    if (normalized === 'deleted') return 'badge badge-error badge-soft'
    return 'badge badge-ghost'
  }

  const totalPages = Math.max(1, Math.ceil(users.length / pageSize))
  const pagedUsers = users.slice((page - 1) * pageSize, page * pageSize)

  return (
    <div className="card bg-base-100 shadow-md">
      <div className={`card-body ${ui.cardBodyPadding}`}>
        <div className="flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <h2 className={`card-title min-w-0 ${ui.cardTitleText}`}>Users</h2>
          <span className={`badge badge-outline shrink-0 ${ui.badgeSize}`}>{users.length} users</span>
        </div>
        {usersError && <div className={`alert alert-error ${ui.tableCellText}`}>{usersError}</div>}
        {isUsersLoading && <div className={`alert alert-info ${ui.tableCellText}`}>Loading users...</div>}
        <div className={`min-w-0 max-w-full ${isMd ? 'max-h-[22rem]' : 'max-h-[28rem]'} overflow-x-auto overflow-y-auto`}>
          <table className={`table table-zebra ${ui.tableSize} min-w-[18rem]`}>
            <thead>
              <tr className={`sticky top-0 z-10 bg-base-100 ${ui.tableCellText}`}>
                <th>User</th>
                {!isMd && <th>Email</th>}
                <th>Status</th>
                <th>Created At</th>
              </tr>
            </thead>
            <tbody>
              {!isUsersLoading && users.length === 0 ? (
                <tr>
                  <td colSpan={isMd ? 3 : 4} className="text-center opacity-70">
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
                    {!isMd && <td>{row.email ?? '-'}</td>}
                    <td>
                      <span className={`${getStatusBadgeClass(row.status)} ${ui.badgeSize}`}>{row.status}</span>
                    </td>
                    <td className="whitespace-nowrap">{formatDateTime(row.createdAt)}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {users.length > 0 && (
          <div className="mt-3 flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <span className={`min-w-0 ${ui.tableCellText}`}>
              Page {page} / {totalPages} - 100 rows per page
            </span>
            <div className="join shrink-0">
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
