import { apiUsersCollectionUrl } from '../api/endpoints'
import UsersTable from '../features/users/UsersTable'

export default function UsersListPage() {
  return <UsersTable usersUrl={apiUsersCollectionUrl} />
}
