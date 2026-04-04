// frontend/src/components/AddExpenseModal.jsx
import { useState } from 'react';
import { api } from '../api.js';

export default function AddExpenseModal({ group, onClose, onSaved }) {
  const [description, setDescription] = useState('');
  const [amount, setAmount] = useState('');
  const [paidBy, setPaidBy] = useState(group.members[0] || '');
  const [splitMode, setSplitMode] = useState('equal');
  const [ratios, setRatios] = useState(() => Object.fromEntries(group.members.map(m => [m, '1'])));
  const [fixed, setFixed] = useState(() => Object.fromEntries(group.members.map(m => [m, '0'])));
  const [error, setError] = useState('');

  function computeSplits() {
    const total = parseFloat(amount);
    if (isNaN(total) || total <= 0) return null;

    if (splitMode === 'equal') {
      const share = total / group.members.length;
      return Object.fromEntries(group.members.map(m => [m, parseFloat(share.toFixed(2))]));
    }

    if (splitMode === 'ratio') {
      const nums = group.members.map(m => parseFloat(ratios[m]) || 0);
      const sum = nums.reduce((a, b) => a + b, 0);
      if (sum === 0) return null;
      return Object.fromEntries(group.members.map((m, i) => [m, parseFloat(((nums[i] / sum) * total).toFixed(2))]));
    }

    return Object.fromEntries(group.members.map(m => [m, parseFloat(fixed[m]) || 0]));
  }

  function splitsTotal() {
    const splits = computeSplits();
    if (!splits) return 0;
    return Object.values(splits).reduce((a, b) => a + b, 0);
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    const splits = computeSplits();
    if (!splits) { setError('Invalid split configuration'); return; }
    const total = parseFloat(amount);
    if (Math.abs(splitsTotal() - total) > 0.01) {
      setError(`Splits sum to ${splitsTotal().toFixed(2)} but total is ${total.toFixed(2)}`);
      return;
    }
    try {
      await api.addExpense(group.id, { description, amount: total, paidBy, splits });
      onSaved();
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <div className="modal-backdrop">
      <div className="modal">
        <h3>Add expense</h3>
        <form onSubmit={handleSubmit}>
          <input placeholder="Description" value={description} onChange={e => setDescription(e.target.value)} required />
          <input type="number" step="0.01" placeholder={`Amount (${group.currency})`} value={amount} onChange={e => setAmount(e.target.value)} required />
          <select value={paidBy} onChange={e => setPaidBy(e.target.value)}>
            {group.members.map(m => <option key={m} value={m}>{m}</option>)}
          </select>
          <select value={splitMode} onChange={e => setSplitMode(e.target.value)}>
            <option value="equal">Equal split</option>
            <option value="ratio">By ratio</option>
            <option value="fixed">Fixed amounts</option>
          </select>

          {splitMode === 'ratio' && group.members.map(m => (
            <div key={m} style={{display:'flex', gap:'0.5rem', alignItems:'center', marginBottom:'0.25rem'}}>
              <label style={{width:'80px'}}>{m}</label>
              <input type="number" step="0.1" value={ratios[m]} onChange={e => setRatios({...ratios, [m]: e.target.value})} style={{width:'80px'}} />
            </div>
          ))}

          {splitMode === 'fixed' && group.members.map(m => (
            <div key={m} style={{display:'flex', gap:'0.5rem', alignItems:'center', marginBottom:'0.25rem'}}>
              <label style={{width:'80px'}}>{m}</label>
              <input type="number" step="0.01" value={fixed[m]} onChange={e => setFixed({...fixed, [m]: e.target.value})} style={{width:'80px'}} />
            </div>
          ))}

          {splitMode !== 'equal' && amount && (
            <p style={{fontSize:'0.85em', color: Math.abs(splitsTotal() - parseFloat(amount)) <= 0.01 ? 'green' : 'red'}}>
              Splits total: {splitsTotal().toFixed(2)} / {parseFloat(amount).toFixed(2)}
            </p>
          )}

          {error && <p className="error">{error}</p>}
          <div style={{display:'flex', gap:'0.5rem', marginTop:'0.5rem'}}>
            <button type="submit">Add</button>
            <button type="button" onClick={onClose}>Cancel</button>
          </div>
        </form>
      </div>
    </div>
  );
}
