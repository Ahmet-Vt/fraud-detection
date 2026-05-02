import { NavLink } from 'react-router-dom'
import arrowIcon from '../assets/arrow.svg'

/** Same shell as the file-tree nav so stats badges align to that width. */
const sidebarNavPanelClass =
  'menu menu-sm w-full rounded-box border border-base-300 bg-base-200 p-2'

type SidebarProps = {
  isMobile: boolean
  connectionMode: 'idle' | 'sse' | 'disconnected'
  requestCount: number
  fraudCount: number
}

function closeMobileMenu(target: HTMLElement) {
  target.closest('details')?.removeAttribute('open')
}

function NavTreeArrow() {
  return (
    <span
      className="inline-flex h-4 w-5 shrink-0 items-center justify-center text-base-content/55"
      aria-hidden
    >
      <img src={arrowIcon} alt="" className="block h-4 w-4" width={16} height={16} decoding="async" />
    </span>
  )
}

function NavButtons() {
  const treeLines: Array<{ to: string; label: string }> = [
    { to: '/livedata', label: 'Live Data' },
    { to: '/users', label: 'Users' },
  ]

  return (
    <nav aria-label="Primary">
      <ul className={sidebarNavPanelClass}>
        {treeLines.map((item) => (
          <li key={item.to}>
            <NavLink
              to={item.to}
              className={({ isActive }) =>
                [
                  'flex items-center gap-2 rounded-lg border border-transparent px-2 py-2 font-sans transition-colors',
                  'hover:border-base-content/15 hover:bg-base-100 hover:shadow-sm',
                  'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary',
                  isActive ? 'active border-primary/30 bg-base-100 font-medium shadow-sm' : '',
                ]
                  .filter(Boolean)
                  .join(' ')
              }
              onClick={(e) => closeMobileMenu(e.currentTarget)}
            >
              <NavTreeArrow />
              <span className="whitespace-nowrap">{item.label}</span>
            </NavLink>
          </li>
        ))}
      </ul>
    </nav>
  )
}

export default function Sidebar({
  isMobile,
  connectionMode,
  requestCount,
  fraudCount,
}: SidebarProps) {
  if (isMobile) {
    return (
      <div className="mb-3 flex justify-end">
        <details className="dropdown dropdown-end">
          <summary className="btn btn-sm btn-boxr" aria-label="Open navigation menu">
            ☰
          </summary>
          <div className="dropdown-content z-[1] mt-2 w-56 rounded-box bg-base-100 p-2 shadow">
            <NavButtons />
          </div>
        </details>
      </div>
    )
  }

  return (
    <aside className="sticky top-0 h-screen w-64 border-r border-base-300 bg-base-100 p-4">
      <h2 className="mb-4 text-lg font-semibold">Menu</h2>
      <NavButtons />
      <div className="mt-6 flex w-full flex-col gap-1 text-sm">
        <div className="badge flex h-auto min-h-0 w-full justify-start border-0 bg-black px-2 py-2 font-normal text-white whitespace-normal">
          Mode: {connectionMode}
        </div>
        <div className="badge flex h-auto min-h-0 w-full justify-start border-0 bg-black px-2 py-2 font-normal text-white whitespace-normal">
          Requests: {requestCount}
        </div>
        <div className="badge flex h-auto min-h-0 w-full justify-start border-0 bg-red-500 px-2 py-2 font-normal text-white whitespace-normal">
          Fraud: {fraudCount}
        </div>
      </div>
    </aside>
  )
}
