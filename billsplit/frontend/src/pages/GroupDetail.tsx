import { useState, useEffect } from 'react'
import { useParams, useOutletContext } from 'react-router-dom'
import { toast } from 'sonner'
import { api, type Group, type GroupEvent } from '@/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import AddExpenseModal from '@/components/AddExpenseModal'
import AddSettlementModal from '@/components/AddSettlementModal'

type OutletContext = { reloadGroups: () => void }

const LIMIT = 20

export default function GroupDetail() {
  const { groupId } = useParams<{ groupId: string }>()
  const { reloadGroups } = useOutletContext<OutletContext>()
  const [group, setGroup] = useState<Group | null>(null)
  const [events, setEvents] = useState<GroupEvent[]>([])
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [showExpense, setShowExpense] = useState(false)
  const [showSettlement, setShowSettlement] = useState(false)

  async function loadGroup() {
    try {
      const g = await api.getGroup(groupId!)
      setGroup(g)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to load group')
    }
  }

  async function loadEvents(newOffset: number) {
    try {
      const res = await api.getExpenses(groupId!, LIMIT, newOffset)
      setEvents(res.events ?? [])
      setTotal(res.total ?? 0)
      setOffset(newOffset)
      setExpandedId(null)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to load expenses')
    }
  }

  useEffect(() => {
    if (groupId) {
      loadGroup()
      loadEvents(0)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [groupId])

  async function handleDeleteExpense(eventId: string) {
    try {
      await api.deleteExpense(groupId!, eventId)
      loadGroup()
      loadEvents(offset)
      reloadGroups()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to cancel expense')
    }
  }

  function refresh() {
    loadGroup()
    loadEvents(0)
    reloadGroups()
  }

  if (!group) return <div className="p-8 text-muted-foreground">Loading...</div>

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold">{group.name}</h1>
          <Badge variant="secondary" className="mt-1">{group.currency}</Badge>
        </div>
        <div className="flex gap-2">
          <Button onClick={() => setShowExpense(true)}>+ Expense</Button>
          <Button variant="outline" onClick={() => setShowSettlement(true)}>+ Settle</Button>
        </div>
      </div>

      {/* Balances */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
            Balances
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {Object.entries(group.balances ?? {}).map(([user, bal]) => (
            <div key={user} className="flex justify-between items-center">
              <span className="text-sm">{user}</span>
              <span className={`text-sm font-semibold ${bal >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                {bal >= 0 ? '+' : ''}{bal.toFixed(2)}
              </span>
            </div>
          ))}
        </CardContent>
      </Card>

      {/* History */}
      <div className="space-y-2">
        <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">History</h2>

        {events.length === 0 && (
          <p className="text-sm text-muted-foreground py-4 text-center">No transactions yet.</p>
        )}

        {events.map(e => (
          <Card key={e.id} className={e.type === 'expense' ? 'cursor-pointer hover:bg-muted/50 transition-colors' : ''}>
            <CardContent className="py-3 px-4">
              {e.type === 'expense' && (
                <div onClick={() => setExpandedId(e.id === expandedId ? null : e.id)}>
                  <div className="flex justify-between items-start">
                    <div>
                      <p className="font-medium text-sm">{e.description}</p>
                      <p className="text-xs text-muted-foreground">
                        Paid by {e.paidBy} · {new Date(e.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                    <div className="text-right ml-4">
                      <p className="font-semibold text-sm">{group.currency} {e.amount.toFixed(2)}</p>
                      <p className="text-xs text-muted-foreground">{expandedId === e.id ? '▲' : '▼'} splits</p>
                    </div>
                  </div>
                  {expandedId === e.id && (
                    <div className="mt-3 pt-3 border-t border-border space-y-1">
                      {Object.entries(e.splits ?? {})
                        .sort(([a], [b]) => a.localeCompare(b))
                        .map(([member, amount]) => (
                          <div key={member} className="flex justify-between text-xs text-muted-foreground max-w-xs">
                            <span>{member}</span>
                            <span>{group.currency} {amount.toFixed(2)}</span>
                          </div>
                        ))}
                      <div className="pt-2">
                        <Button
                          size="sm"
                          variant="destructive"
                          className="text-xs h-7"
                          onClick={ev => { ev.stopPropagation(); handleDeleteExpense(e.id) }}
                        >
                          Cancel expense
                        </Button>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {e.type === 'settlement' && (
                <div className="flex justify-between items-center">
                  <div>
                    <p className="font-medium text-sm">Settlement</p>
                    <p className="text-xs text-muted-foreground">
                      {e.from} → {e.to} · {new Date(e.createdAt).toLocaleDateString()}
                    </p>
                  </div>
                  <p className="font-semibold text-sm">{group.currency} {e.amount.toFixed(2)}</p>
                </div>
              )}

              {e.type === 'reversal' && (
                <p className="text-xs text-muted-foreground italic">
                  Expense cancelled (#{e.reversedEventId?.slice(0, 8)})
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Pagination */}
      {(offset > 0 || offset + LIMIT < total) && (
        <div className="flex gap-2">
          {offset > 0 && (
            <Button variant="outline" size="sm" onClick={() => loadEvents(offset - LIMIT)}>← Prev</Button>
          )}
          {offset + LIMIT < total && (
            <Button variant="outline" size="sm" onClick={() => loadEvents(offset + LIMIT)}>Next →</Button>
          )}
        </div>
      )}

      {showExpense && (
        <AddExpenseModal
          group={group}
          onClose={() => setShowExpense(false)}
          onSaved={() => { setShowExpense(false); refresh() }}
        />
      )}
      {showSettlement && (
        <AddSettlementModal
          group={group}
          onClose={() => setShowSettlement(false)}
          onSaved={() => { setShowSettlement(false); refresh() }}
        />
      )}
    </div>
  )
}
