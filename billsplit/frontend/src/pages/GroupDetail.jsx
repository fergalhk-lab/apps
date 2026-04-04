// frontend/src/pages/GroupDetail.jsx
import { useState, useEffect } from 'react';
import { api } from '../api.js';
import AddExpenseModal from '../components/AddExpenseModal.jsx';
import AddSettlementModal from '../components/AddSettlementModal.jsx';

export default function GroupDetail({ groupId, onBack, onLogout }) {
  const [group, setGroup] = useState(null);
  const [events, setEvents] = useState([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState('');
  const [showExpense, setShowExpense] = useState(false);
  const [showSettlement, setShowSettlement] = useState(false);
  const [expandedId, setExpandedId] = useState(null);
  const LIMIT = 20;

  useEffect(() => { loadGroup(); loadEvents(0); }, [groupId]);

  async function loadGroup() {
    try {
      const g = await api.getGroup(groupId);
      setGroup(g);
    } catch (err) {
      setError(err.message);
    }
  }

  async function loadEvents(newOffset) {
    try {
      const res = await api.getExpenses(groupId, LIMIT, newOffset);
      setEvents(res.events || []);
      setTotal(res.total || 0);
      setOffset(newOffset);
      setExpandedId(null);
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleDeleteExpense(eventId) {
    try {
      await api.deleteExpense(groupId, eventId);
      loadGroup();
      loadEvents(offset);
    } catch (err) {
      setError(err.message);
    }
  }

  function refresh() {
    loadGroup();
    loadEvents(0);
  }

  if (!group) return <div>{error || 'Loading...'}</div>;

  return (
    <div>
      <div style={{display:'flex', justifyContent:'space-between', marginBottom:'1rem'}}>
        <button onClick={onBack}>← Back</button>
        <button onClick={onLogout}>Sign out</button>
      </div>
      <h1>{group.name} <span style={{color:'#666', fontSize:'0.8em'}}>({group.currency})</span></h1>

      <h2 style={{marginTop:'1rem'}}>Balances</h2>
      {Object.entries(group.balances || {}).map(([user, bal]) => (
        <div key={user} style={{color: bal >= 0 ? 'green' : 'red'}}>
          {user}: {bal >= 0 ? `owed ${bal.toFixed(2)}` : `owes ${Math.abs(bal).toFixed(2)}`}
        </div>
      ))}

      <div style={{display:'flex', gap:'0.5rem', margin:'1rem 0'}}>
        <button onClick={() => setShowExpense(true)}>+ Expense</button>
        <button onClick={() => setShowSettlement(true)}>+ Settlement</button>
      </div>

      {error && <p className="error">{error}</p>}

      <h2>History</h2>
      {events.map(e => (
        <div
          key={e.id}
          className="card"
          onClick={() => e.type === 'expense' && setExpandedId(e.id === expandedId ? null : e.id)}
          style={e.type === 'expense' ? {cursor:'pointer'} : undefined}
        >
          {e.type === 'expense' && (
            <>
              <strong>{e.description}</strong> — {group.currency} {e.amount.toFixed(2)} paid by {e.paidBy}
              <div style={{fontSize:'0.85em', color:'#666'}}>
                {new Date(e.createdAt).toLocaleDateString()} {expandedId === e.id ? '▲' : '▼'}
              </div>
              <button
                onClick={ev => { ev.stopPropagation(); handleDeleteExpense(e.id); }}
                style={{marginTop:'0.25rem', fontSize:'0.8em'}}
              >Cancel</button>
              {expandedId === e.id && (
                <div style={{marginTop:'0.5rem', fontSize:'0.85em', borderTop:'1px solid #eee', paddingTop:'0.5rem'}}>
                  <strong>Splits:</strong>
                  {Object.entries(e.splits || {})
                    .sort(([a], [b]) => a.localeCompare(b))
                    .map(([member, amount]) => (
                      <div key={member} style={{display:'flex', justifyContent:'space-between', maxWidth:'200px'}}>
                        <span>{member}</span>
                        <span>{group.currency} {amount.toFixed(2)}</span>
                      </div>
                    ))
                  }
                </div>
              )}
            </>
          )}
          {e.type === 'settlement' && (
            <>
              <strong>Settlement</strong>: {e.from} paid {e.to} {group.currency} {e.amount.toFixed(2)}
              <div style={{fontSize:'0.85em', color:'#666'}}>{new Date(e.createdAt).toLocaleDateString()}</div>
            </>
          )}
          {e.type === 'reversal' && (
            <span style={{color:'#999'}}>Expense cancelled (#{e.reversedEventId.slice(0,8)})</span>
          )}
        </div>
      ))}

      <div style={{display:'flex', gap:'0.5rem', marginTop:'1rem'}}>
        {offset > 0 && <button onClick={() => loadEvents(offset - LIMIT)}>← Prev</button>}
        {offset + LIMIT < total && <button onClick={() => loadEvents(offset + LIMIT)}>Next →</button>}
      </div>

      {showExpense && (
        <AddExpenseModal
          group={group}
          onClose={() => setShowExpense(false)}
          onSaved={() => { setShowExpense(false); refresh(); }}
        />
      )}
      {showSettlement && (
        <AddSettlementModal
          group={group}
          onClose={() => setShowSettlement(false)}
          onSaved={() => { setShowSettlement(false); refresh(); }}
        />
      )}
    </div>
  );
}
