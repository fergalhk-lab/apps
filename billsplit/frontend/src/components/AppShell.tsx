// frontend-new/src/components/AppShell.tsx
import { useEffect, useState } from 'react'
import { Link, Outlet, useMatch, useNavigate } from 'react-router-dom'
import { Sun, Moon, Monitor, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { api, clearToken, parseToken, TOKEN_KEY, type Group } from '@/api'
import { useTheme } from '@/components/ThemeProvider'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import CreateGroupDialog from '@/components/CreateGroupDialog'

export default function AppShell() {
  const [groups, setGroups] = useState<Group[]>([])
  const [showCreateGroup, setShowCreateGroup] = useState(false)
  const [showInvitePanel, setShowInvitePanel] = useState(false)
  const [inviteIsAdmin, setInviteIsAdmin] = useState(false)
  const [inviteCode, setInviteCode] = useState('')
  const { theme, setTheme } = useTheme()
  const navigate = useNavigate()
  const match = useMatch('/groups/:groupId')
  const activeGroupId = match?.params.groupId ?? null

  const token = localStorage.getItem(TOKEN_KEY) ?? ''
  const { username, isAdmin } = parseToken(token)

  useEffect(() => {
    loadGroups()
  }, [])

  async function loadGroups() {
    try {
      const list = await api.getGroups()
      setGroups(list ?? [])
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to load groups')
    }
  }

  function handleLogout() {
    clearToken()
    navigate('/login')
  }

  async function handleGenerateInvite() {
    try {
      const res = await api.generateInvite(inviteIsAdmin)
      setInviteCode(res.code)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to generate invite')
    }
  }

  function cycleTheme() {
    if (theme === 'system') setTheme('light')
    else if (theme === 'light') setTheme('dark')
    else setTheme('system')
  }

  const ThemeIcon = theme === 'light' ? Sun : theme === 'dark' ? Moon : Monitor

  return (
    <div className="flex h-screen bg-background text-foreground">
      {/* Sidebar */}
      <aside className="w-56 flex-shrink-0 border-r border-border flex flex-col">
        {/* Logo */}
        <div className="px-4 py-4">
          <span className="font-bold text-lg">BillSplit</span>
        </div>

        <Separator />

        {/* Groups list */}
        <div className="flex-1 overflow-y-auto py-2">
          <p className="px-4 py-1 text-xs font-medium text-muted-foreground uppercase tracking-wider">
            Groups
          </p>
          {groups.map(g => (
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
          ))}

          {/* New group button */}
          <button
            onClick={() => setShowCreateGroup(true)}
            className="flex items-center gap-1 px-4 py-2 mx-2 mt-1 text-sm text-muted-foreground hover:text-foreground hover:bg-muted rounded-md w-full transition-colors"
          >
            <Plus className="h-3 w-3" />
            New group
          </button>
        </div>

        <Separator />

        {/* Admin invite panel */}
        {isAdmin && (
          <div className="px-4 py-2">
            <button
              onClick={() => { setShowInvitePanel(v => !v); setInviteCode('') }}
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
                    onChange={e => setInviteIsAdmin(e.target.checked)}
                  />
                  Admin invite
                </label>
                <Button size="sm" variant="outline" className="w-full text-xs" onClick={handleGenerateInvite}>
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
                      onClick={() =>
                        navigator.clipboard.writeText(inviteCode)
                          .then(() => toast.success('Copied!'))
                          .catch(() => toast.error('Failed to copy'))
                      }
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
            onClick={handleLogout}
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

      {/* Main content */}
      <main className="flex-1 overflow-y-auto">
        <Outlet context={{ reloadGroups: loadGroups }} />
      </main>

      {showCreateGroup && (
        <CreateGroupDialog
          currentUsername={username}
          onClose={() => setShowCreateGroup(false)}
          onCreated={(groupId) => {
            setShowCreateGroup(false)
            loadGroups()
            navigate(`/groups/${groupId}`)
          }}
        />
      )}
    </div>
  )
}
