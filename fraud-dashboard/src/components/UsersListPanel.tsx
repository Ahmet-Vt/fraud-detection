import { apiUsersCollectionUrl } from '../api/endpoints'
import UsersTable from './UsersTable'

export default function UsersListPanel() {
  return <UsersTable usersUrl={apiUsersCollectionUrl} />
}
