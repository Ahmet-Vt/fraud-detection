type AppPage = 'live-data' | 'users'

type SidebarProps = {
  isMobile: boolean
  activePage: AppPage
  onNavigate: (page: AppPage) => void
  connectionMode: 'idle' | 'sse' | 'disconnected'
  requestCount: number
  fraudCount: number
}

const navItems: Array<{ page: AppPage; label: string }> = [
  { page: 'live-data', label: 'Live Data' },
  { page: 'users', label: 'Users' },
]

function NavButtons({
  activePage,
  onNavigate,
}: {
  activePage: AppPage
  onNavigate: (page: AppPage) => void
}) {
  return (
    <ul className="menu gap-1">
      {navItems.map((item) => (
        <li key={item.page}>
          <button
            type="button"
            className={activePage === item.page ? 'active font-medium' : ''}
            onClick={() => onNavigate(item.page)}
          >
            {item.label}
          </button>
        </li>
      ))}
    </ul>
  )
}

export default function Sidebar({
  isMobile,
  activePage,
  onNavigate,
  connectionMode,
  requestCount,
  fraudCount,
}: SidebarProps) {
  if (isMobile) {
    return (
      <div className="mb-3 flex justify-end">
        <details className="dropdown dropdown-end">
          <summary className="btn btn-sm btn-outline" aria-label="Open navigation menu">
            ☰
          </summary>
          <div className="dropdown-content z-[1] mt-2 w-56 rounded-box bg-base-100 p-2 shadow">
            <NavButtons activePage={activePage} onNavigate={onNavigate} />
          </div>
        </details>
      </div>
    )
  }

  return (
    <aside className="sticky top-0 h-screen w-64 border-r border-base-300 bg-base-100 p-4">
      <h2 className="mb-4 text-lg font-semibold">Menu</h2>
      <NavButtons activePage={activePage} onNavigate={onNavigate} />
      <div className="mt-6 space-y-2 text-sm">
        <div className="badge badge-outline w-full justify-start">Mode: {connectionMode}</div>
        <div className="badge badge-info badge-outline w-full justify-start">Requests: {requestCount}</div>
        <div className="badge badge-error badge-outline w-full justify-start">Fraud: {fraudCount}</div>
      </div>
    </aside>
  )
}
