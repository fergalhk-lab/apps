// frontend/src/components/CreateGroupDialog.tsx
import { useState, useEffect } from 'react'
import { api, type UserSummary } from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { X } from 'lucide-react'

interface Props {
  currentUsername: string
  onClose: () => void
  onCreated: (groupId: string) => void
}

export default function CreateGroupDialog({ currentUsername, onClose, onCreated }: Props) {
  const [name, setName] = useState('')
  const [currency, setCurrency] = useState('GBP')
  const [availableUsers, setAvailableUsers] = useState<UserSummary[]>([])
  const [selectedMembers, setSelectedMembers] = useState<UserSummary[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    api.getUsers()
      .then(res => setAvailableUsers((res.users ?? []).filter(u => u.id !== currentUsername)))
      .catch(err => setError(err instanceof Error ? err.message : 'Failed to load users'))
  }, [currentUsername])

  function toggleMember(user: UserSummary) {
    setSelectedMembers(prev =>
      prev.some(m => m.id === user.id)
        ? prev.filter(m => m.id !== user.id)
        : [...prev, user]
    )
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    if (selectedMembers.length === 0) {
      setError('Add at least one other member')
      return
    }
    try {
      const group = await api.createGroup(name, currency, selectedMembers.map(u => u.id))
      onCreated(group.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create group')
    }
  }

  return (
    <Dialog open onOpenChange={open => !open && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create group</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Group name</Label>
            <Input
              placeholder="Weekend trip"
              value={name}
              onChange={e => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Currency</Label>
            <Input
              placeholder="GBP"
              value={currency}
              onChange={e => setCurrency(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Members</Label>
            {/* Selected members chips */}
            <div className="min-h-10 border rounded-md p-2 flex flex-wrap gap-1">
              {selectedMembers.length === 0 && (
                <span className="text-xs text-muted-foreground italic self-center">
                  Click below to add members…
                </span>
              )}
              {selectedMembers.map(u => (
                <Badge key={u.id} variant="secondary" className="flex items-center gap-1">
                  {u.id}
                  <button
                    type="button"
                    onClick={() => toggleMember(u)}
                    className="ml-1 hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              ))}
            </div>
            {/* Available users list */}
            <div className="max-h-36 overflow-y-auto border rounded-md divide-y">
              {availableUsers.map(u => {
                const selected = selectedMembers.some(m => m.id === u.id)
                return (
                  <button
                    key={u.id}
                    type="button"
                    onClick={() => toggleMember(u)}
                    className={`w-full flex items-center gap-2 px-3 py-2 text-sm text-left transition-colors ${
                      selected ? 'bg-primary/10 text-primary' : 'hover:bg-muted'
                    }`}
                  >
                    <span className="w-4 text-center">{selected ? '✓' : ''}</span>
                    {u.id}
                  </button>
                )
              })}
              {availableUsers.length === 0 && !error && (
                <p className="text-xs text-muted-foreground p-3">No other users found.</p>
              )}
            </div>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex gap-2 justify-end">
            <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
            <Button type="submit">Create</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
