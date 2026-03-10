import { NavLink, Outlet } from 'react-router-dom'

const navItems = [
  ['/', '总览'],
  ['/comics', '漫画'],
  ['/favorites', '收藏'],
  ['/history', '历史'],
  ['/sources', '源站'],
  ['/settings', '设置'],
  ['/downloads', '下载'],
]

export default function Layout() {
  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">合</div>
          <div>
            <strong>合欢阅读器</strong>
            <p>本地优先 · 多源容灾</p>
          </div>
        </div>
        <nav className="nav">
          {navItems.map(([to, label]) => (
            <NavLink key={`${to}-${label}`} to={to} className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')}>
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="content">
        <Outlet />
      </main>
    </div>
  )
}
