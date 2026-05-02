import UsersTable from '../components/UsersTable'

export default function UsersListPage() {
  const env = import.meta.env as Record<string, string | undefined>
  const usersUrl = env.VITE_USERS_URL ?? '/users'
  return <UsersTable usersUrl={usersUrl} />
}
