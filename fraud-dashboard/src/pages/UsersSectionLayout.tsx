import { Outlet } from 'react-router-dom'
import { useScreen } from '../config/ScreenContext'

export default function UsersSectionLayout() {
  const { ui } = useScreen()
  return (
    <div className={`space-y-3 ${ui.stackGap}`}>
      <Outlet />
    </div>
  )
}
