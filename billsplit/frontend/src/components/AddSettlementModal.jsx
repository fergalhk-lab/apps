// frontend/src/components/AddSettlementModal.jsx
import { useState } from 'react';
import { api } from '../api.js';

export default function AddSettlementModal({ group, onClose, onSaved }) {
  const [from, setFrom] = useState(group.members[0] || '');
  const [to, setTo] = useState(group.members[1] || group.members[0] || '');
  const [amount, setAmount] = useState('');
  const [error, setError] = useState('');

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    try {
      await api.addSettlement(group.id, { from, to, amount: parseFloat(amount) });
      onSaved();
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <div className="modal-backdrop">
      <div className="modal">
        <h3>Record settlement</h3>
        <form onSubmit={handleSubmit}>
          <label>From</label>
          <select value={from} onChange={e => setFrom(e.target.value)}>
            {group.members.map(m => <option key={m} value={m}>{m}</option>)}
          </select>
          <label>To</label>
          <select value={to} onChange={e => setTo(e.target.value)}>
            {group.members.map(m => <option key={m} value={m}>{m}</option>)}
          </select>
          <input type="number" step="0.01" placeholder={`Amount (${group.currency})`} value={amount} onChange={e => setAmount(e.target.value)} required />
          {error && <p className="error">{error}</p>}
          <div style={{display:'flex', gap:'0.5rem', marginTop:'0.5rem'}}>
            <button type="submit">Record</button>
            <button type="button" onClick={onClose}>Cancel</button>
          </div>
        </form>
      </div>
    </div>
  );
}
