import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScreen } from '../../config/ScreenContext'
import { formatDateTime } from '../../lib/formatDateTime'
import { AppCard, InlineAlert, SimplePagination, tableScrollAreaClass } from '../../components/ui/dashboardUi'
import { useUsersList } from './hooks/useUsersList'

export type { UserRow } from './userTypes'

type UsersTableProps = {
  usersUrl: string
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
    <AppCard bodyPaddingClass={ui.cardBodyPadding}>
      <div className="flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <h2 className={`card-title min-w-0 ${ui.cardTitleText}`}>Users</h2>
        <span className={`badge badge-outline shrink-0 ${ui.badgeSize}`}>{users.length} users</span>
      </div>
      {usersError && <InlineAlert variant="error" textClass={ui.tableCellText}>{usersError}</InlineAlert>}
      {isUsersLoading && <InlineAlert variant="info" textClass={ui.tableCellText}>Loading users...</InlineAlert>}
      <div className={tableScrollAreaClass(isMd)}>
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
        <SimplePagination
          page={page}
          totalPages={totalPages}
          isMd={isMd}
          textClass={ui.tableCellText}
          pageSize={pageSize}
          onPrev={() => setPage((p) => Math.max(1, p - 1))}
          onNext={() => setPage((p) => Math.min(totalPages, p + 1))}
        />
      )}
    </AppCard>
  )
}
