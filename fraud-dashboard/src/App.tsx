import { Navigate, Route, Routes } from 'react-router-dom'
import DashboardLayout from './layout/DashboardLayout'
import LiveDataPage from './pages/LiveDataPage'
import UserDetailPage from './pages/UserDetailPage'
import UsersListPage from './pages/UsersListPage'
import UsersSectionLayout from './pages/UsersSectionLayout'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<DashboardLayout />}>
        <Route index element={<Navigate to="/livedata" replace />} />
        <Route path="livedata" element={<LiveDataPage />} />
        <Route path="users" element={<UsersSectionLayout />}>
          <Route index element={<UsersListPage />} />
          <Route path=":userId" element={<UserDetailPage />} />
        </Route>
      </Route>
    </Routes>
  )
}
