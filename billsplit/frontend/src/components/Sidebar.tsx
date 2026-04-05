// frontend/src/components/Sidebar.tsx
import { Link } from 'react-router-dom'
import { Sun, Moon, Monitor, Plus, ChevronLeft, ChevronRight } from 'lucide-react'
import { type Group } from '@/api'
import { useTheme } from '@/components/ThemeProvider'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { groupInitial } from '@/components/sidebar-utils'

interface SidebarProps {
  groups: Group[]
  activeGroupId: string | null
  username: string
  isAdmin: boolean
  collapsed: boolean
  onToggle: () => void
  onNewGroup: () => void
  onLogout: () => void
  showInvitePanel: boolean
  onToggleInvitePanel: () => void
  inviteIsAdmin: boolean
  onSetInviteIsAdmin: (v: boolean) => void
  inviteCode: string
  onGenerateInvite: () => void
  onCopyInvite: () => void
}

export default function Sidebar({
  groups,
  activeGroupId,
  username,
  isAdmin,
  collapsed,
  onToggle,
  onNewGroup,
  onLogout,
  showInvitePanel,
  onToggleInvitePanel,
  inviteIsAdmin,
  onSetInviteIsAdmin,
  inviteCode,
  onGenerateInvite,
  onCopyInvite,
}: SidebarProps) {
  const { theme, setTheme } = useTheme()

  function cycleTheme() {
    if (theme === 'system') setTheme('light')
    else if (theme === 'light') setTheme('dark')
    else setTheme('system')
  }

  const ThemeIcon = theme === 'light' ? Sun : theme === 'dark' ? Moon : Monitor

  return (
    <aside
      className={`flex-shrink-0 border-r border-border flex flex-col transition-all duration-200 ${
        collapsed ? 'w-11' : 'w-56'
      }`}
    >
      {/* Toggle + Logo */}
      <div className={`flex items-center py-3 ${collapsed ? 'justify-center px-0' : 'justify-between px-4'}`}>
        {!collapsed && <span className="font-bold text-lg">BillSplit</span>}
        <button
          onClick={onToggle}
          className="text-muted-foreground hover:text-foreground transition-colors p-1 rounded"
          title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          aria-expanded={!collapsed}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
        </button>
      </div>

      {/* Groups list */}
      <div className="flex-1 overflow-y-auto py-2">
        {!collapsed && (
          <p className="px-4 py-1 text-xs font-medium text-muted-foreground uppercase tracking-wider">
            Groups
          </p>
        )}

        {groups.map(g =>
          collapsed ? (
            <Link
              key={g.id}
              to={`/groups/${g.id}`}
              title={g.name}
              aria-label={g.name}
              className="flex items-center justify-center py-1 mx-1 my-0.5"
            >
              <span
                className={`w-7 h-7 rounded-md flex items-center justify-center text-xs font-bold text-white transition-all ${
                  g.id === activeGroupId
                    ? 'bg-primary ring-2 ring-primary ring-offset-1'
                    : 'bg-muted-foreground/50'
                }`}
              >
                {groupInitial(g.name)}
              </span>
            </Link>
          ) : (
            <Link
              key={g.id}
              to={`/groups/${g.id}`}
              className={`flex items-center justify-between px-4 py-2 text-sm rounded-md mx-2 my-0.5 transition-colors ${
                g.id === activeGroupId
                  ? 'bg-primary/10 text-primary font-medium'
                  : 'text-foreground hover:bg-muted'
              }`}
            >
              <span className="truncate">{g.name}</span>
              <span
                className={`text-xs font-medium ml-2 flex-shrink-0 ${
                  g.netBalance >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
                }`}
              >
                {g.netBalance >= 0 ? '+' : ''}{g.netBalance.toFixed(2)}
              </span>
            </Link>
          )
        )}

        {/* New group button */}
        {collapsed ? (
          <button
            onClick={onNewGroup}
            title="New group"
            aria-label="New group"
            className="flex items-center justify-center py-1 mx-1 my-0.5 w-full text-muted-foreground hover:text-foreground transition-colors"
          >
            <Plus className="h-4 w-4" />
          </button>
        ) : (
          <button
            onClick={onNewGroup}
            className="flex items-center gap-1 px-4 py-2 mx-2 mt-1 text-sm text-muted-foreground hover:text-foreground hover:bg-muted rounded-md w-full transition-colors"
          >
            <Plus className="h-3 w-3" />
            New group
          </button>
        )}
      </div>

      <Separator />

      {/* Admin invite panel */}
      {isAdmin && (
        <div className="px-4 py-2">
          <button
            onClick={onToggleInvitePanel}
            className="text-xs text-muted-foreground hover:text-foreground transition-colors"
          >
            Generate invite
          </button>
          {showInvitePanel && (
            <div className="mt-2 space-y-2">
              <label className="flex items-center gap-2 text-xs">
                <input
                  type="checkbox"
                  checked={inviteIsAdmin}
                  onChange={e => onSetInviteIsAdmin(e.target.checked)}
                />
                Admin invite
              </label>
              <Button size="sm" variant="outline" className="w-full text-xs" onClick={onGenerateInvite}>
                Generate
              </Button>
              {inviteCode && (
                <div className="flex gap-1">
                  <input
                    readOnly
                    value={inviteCode}
                    className="flex-1 text-xs border rounded px-2 py-1 bg-muted font-mono"
                  />
                  <Button
                    size="sm"
                    variant="outline"
                    className="text-xs px-2"
                    onClick={onCopyInvite}
                  >
                    Copy
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Bottom: username + theme toggle */}
      <div className="px-4 py-3 flex items-center justify-between">
        <button
          onClick={onLogout}
          className="text-sm text-muted-foreground hover:text-foreground transition-colors truncate"
        >
          {username}
        </button>
        <button
          onClick={cycleTheme}
          className="text-muted-foreground hover:text-foreground transition-colors ml-2 flex-shrink-0"
          title={`Theme: ${theme}`}
        >
          <ThemeIcon className="h-4 w-4" />
        </button>
      </div>
    </aside>
  )
}
