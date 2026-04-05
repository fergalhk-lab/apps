// frontend/src/components/AddExpenseModal.tsx
import { useState, useEffect } from 'react'
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

type SplitMode = 'equal' | 'ratio' | 'fixed'

// Exported for testing
export function computeSplits(
  mode: SplitMode,
  total: number,
  members: string[],
  ratios: Record<string, string>,
  fixed: Record<string, string>,
): Record<string, number> | null {
  if (isNaN(total) || total <= 0) return null

  if (mode === 'equal') {
    const share = parseFloat((total / members.length).toFixed(2))
    return Object.fromEntries(members.map(m => [m, share]))
  }

  if (mode === 'ratio') {
    const nums = members.map(m => parseFloat(ratios[m]) || 0)
    const sum = nums.reduce((a, b) => a + b, 0)
    if (sum === 0) return null
    return Object.fromEntries(
      members.map((m, i) => [m, parseFloat(((nums[i] / sum) * total).toFixed(2))])
    )
  }

  return Object.fromEntries(members.map(m => [m, parseFloat(fixed[m]) || 0]))
}

interface Props {
  group: Group
  currentUsername: string
  onClose: () => void
  onSaved: () => void
}

export default function AddExpenseModal({ group, currentUsername, onClose, onSaved }: Props) {
  const [description, setDescription] = useState('')
  const [amount, setAmount] = useState('')
  const [currency, setCurrency] = useState(group.currency)
  const [currencies, setCurrencies] = useState<string[]>([])
  const [paidBy, setPaidBy] = useState(
    group.members.includes(currentUsername) ? currentUsername : (group.members[0] ?? '')
  )
  const [splitMode, setSplitMode] = useState<SplitMode>('equal')
  const [ratios, setRatios] = useState<Record<string, string>>(
    () => Object.fromEntries(group.members.map(m => [m, '1']))
  )
  const [fixed, setFixed] = useState<Record<string, string>>(
    () => Object.fromEntries(group.members.map(m => [m, '0']))
  )
  const [error, setError] = useState('')

  useEffect(() => {
    api.getCurrencies()
      .then(res => setCurrencies(Object.keys(res.rates).sort()))
      .catch(() => {/* silently fall back to group currency only */})
  }, [])

  const total = parseFloat(amount)
  const splits = computeSplits(splitMode, total, group.members, ratios, fixed)
  const splitsTotal = splits ? Object.values(splits).reduce((a, b) => a + b, 0) : 0
  const splitsMismatch = splitMode === 'fixed' && amount && splits && Math.abs(splitsTotal - total) > 0.01

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    if (!splits) { setError('Invalid split configuration'); return }
    if (splitsMismatch) {
      setError(`Splits sum to ${splitsTotal.toFixed(2)} but total is ${total.toFixed(2)}`)
      return
    }
    try {
      await api.addExpense(group.id, { description, amount: total, paidBy, splits, currency })
      onSaved()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add expense')
    }
  }

  return (
    <Dialog open onOpenChange={open => !open && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add expense</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Description</Label>
            <Input
              placeholder="Description"
              value={description}
              onChange={e => setDescription(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Amount</Label>
            <div className="flex gap-2">
              <Input
                type="number"
                step="0.01"
                min="0.01"
                placeholder="0.00"
                value={amount}
                onChange={e => setAmount(e.target.value)}
                required
              />
              <Select value={currency} onValueChange={setCurrency}>
                <SelectTrigger className="w-28">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {(currencies.length > 0 ? currencies : [group.currency]).map(c => (
                    <SelectItem key={c} value={c}>{c}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="space-y-2">
            <Label>Paid by</Label>
            <Select value={paidBy} onValueChange={setPaidBy}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {group.members.map(m => <SelectItem key={m} value={m}>{m}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Split</Label>
            <Select value={splitMode} onValueChange={v => setSplitMode(v as SplitMode)}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="equal">Equal split</SelectItem>
                <SelectItem value="ratio">By ratio</SelectItem>
                <SelectItem value="fixed">Fixed amounts</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {splitMode === 'ratio' && (
            <div className="space-y-2">
              {group.members.map(m => (
                <div key={m} className="flex items-center gap-3">
                  <Label className="w-24 truncate">{m}</Label>
                  <Input
                    type="number"
                    step="0.1"
                    min="0"
                    className="w-24"
                    value={ratios[m]}
                    onChange={e => setRatios({ ...ratios, [m]: e.target.value })}
                  />
                </div>
              ))}
            </div>
          )}

          {splitMode === 'fixed' && (
            <div className="space-y-2">
              {group.members.map(m => (
                <div key={m} className="flex items-center gap-3">
                  <Label className="w-24 truncate">{m}</Label>
                  <Input
                    type="number"
                    step="0.01"
                    min="0"
                    className="w-24"
                    value={fixed[m]}
                    onChange={e => setFixed({ ...fixed, [m]: e.target.value })}
                  />
                </div>
              ))}
              {amount && (
                <p className={`text-xs ${splitsMismatch ? 'text-destructive' : 'text-green-600 dark:text-green-400'}`}>
                  Splits total: {splitsTotal.toFixed(2)} / {total.toFixed(2)}
                </p>
              )}
            </div>
          )}

          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex gap-2 justify-end">
            <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
            <Button type="submit">Add</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
