// frontend/src/components/AppShell.tsx
import { useEffect, useState } from 'react'
import { Outlet, useMatch, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { api, clearToken, parseToken, TOKEN_KEY, type Group } from '@/api'
import Sidebar from '@/components/Sidebar'
import CreateGroupDialog from '@/components/CreateGroupDialog'

export default function AppShell() {
  const [groups, setGroups] = useState<Group[]>([])
  const [showCreateGroup, setShowCreateGroup] = useState(false)
  const [showInvitePanel, setShowInvitePanel] = useState(false)
  const [inviteIsAdmin, setInviteIsAdmin] = useState(false)
  const [inviteCode, setInviteCode] = useState('')
  const [collapsed, setCollapsed] = useState(
    () => localStorage.getItem('sidebar-collapsed') === 'true'
  )
  const navigate = useNavigate()
  const match = useMatch('/groups/:groupId')
  const activeGroupId = match?.params.groupId ?? null

  const token = localStorage.getItem(TOKEN_KEY) ?? ''
  const { username, isAdmin } = parseToken(token)

  async function loadGroups() {
    try {
      const list = await api.getGroups()
      setGroups(list ?? [])
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to load groups')
    }
  }

  useEffect(() => {
    loadGroups()
  }, []) // loadGroups is intentionally excluded — initial load only

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

  function handleCopyInvite() {
    navigator.clipboard.writeText(inviteCode)
      .then(() => toast.success('Copied!'))
      .catch(() => toast.error('Failed to copy'))
  }

  return (
    <div className="flex h-screen bg-background text-foreground">
      <Sidebar
        groups={groups}
        activeGroupId={activeGroupId}
        username={username}
        isAdmin={isAdmin}
        collapsed={collapsed}
        onToggle={() => setCollapsed(v => {
          const next = !v
          localStorage.setItem('sidebar-collapsed', String(next))
          if (next) { setShowInvitePanel(false); setInviteCode('') }
          return next
        })}
        onNewGroup={() => setShowCreateGroup(true)}
        onLogout={handleLogout}
        showInvitePanel={showInvitePanel}
        onToggleInvitePanel={() => { setShowInvitePanel(v => !v); setInviteCode('') }}
        inviteIsAdmin={inviteIsAdmin}
        onSetInviteIsAdmin={setInviteIsAdmin}
        inviteCode={inviteCode}
        onGenerateInvite={handleGenerateInvite}
        onCopyInvite={handleCopyInvite}
      />

      <main className="flex-1 overflow-y-auto">
        <Outlet context={{ reloadGroups: loadGroups, username }} />
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
