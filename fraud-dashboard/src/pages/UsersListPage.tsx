import { apiUsersCollectionUrl } from '../api/endpoints'
import UsersTable from '../components/UsersTable'

export default function UsersListPage() {
  return <UsersTable usersUrl={apiUsersCollectionUrl} />
}
