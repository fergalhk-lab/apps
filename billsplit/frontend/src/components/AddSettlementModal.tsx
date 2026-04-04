// frontend/src/components/AddSettlementModal.tsx
import { useState } from 'react'
import { api, type Group } from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface Props {
  group: Group
  onClose: () => void
  onSaved: () => void
}

export default function AddSettlementModal({ group, onClose, onSaved }: Props) {
  const [from, setFrom] = useState(group.members[0] ?? '')
  const [to, setTo] = useState(group.members[1] ?? group.members[0] ?? '')
  const [amount, setAmount] = useState('')
  const [error, setError] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    if (from === to) {
      setError('Cannot settle with yourself')
      return
    }
    try {
      await api.addSettlement(group.id, { from, to, amount: parseFloat(amount) })
      onSaved()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to record settlement')
    }
  }

  return (
    <Dialog open onOpenChange={open => !open && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Record settlement</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>From</Label>
            <Select value={from} onValueChange={setFrom}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {group.members.map(m => <SelectItem key={m} value={m}>{m}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>To</Label>
            <Select value={to} onValueChange={setTo}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {group.members.map(m => <SelectItem key={m} value={m}>{m}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Amount ({group.currency})</Label>
            <Input
              type="number"
              step="0.01"
              min="0.01"
              placeholder="0.00"
              value={amount}
              onChange={e => setAmount(e.target.value)}
              required
            />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex gap-2 justify-end">
            <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
            <Button type="submit">Record</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
