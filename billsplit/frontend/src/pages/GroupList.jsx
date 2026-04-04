// frontend/src/pages/GroupList.jsx
import { useState, useEffect } from 'react';
import { api, parseToken } from '../api.js';

export default function GroupList({ onSelectGroup, onLogout, isAdmin, token }) {
  const [groups, setGroups] = useState([]);
  const [error, setError] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [name, setName] = useState('');
  const [currency, setCurrency] = useState('GBP');
  const [selectedMembers, setSelectedMembers] = useState([]);
  const [availableUsers, setAvailableUsers] = useState([]);
  const [createError, setCreateError] = useState('');

  const [showInvite, setShowInvite] = useState(false);
  const [inviteIsAdmin, setInviteIsAdmin] = useState(false);
  const [inviteCode, setInviteCode] = useState('');
  const [inviteError, setInviteError] = useState('');

  useEffect(() => { loadGroups(); }, []);

  async function loadGroups() {
    try {
      const list = await api.getGroups();
      setGroups(list || []);
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleShowCreate() {
    try {
      const res = await api.getUsers();
      const activeUsername = parseToken(token).username;
      // u.id is the username string (see domain.UserSummary)
      setAvailableUsers((res.users || []).filter(u => u.id !== activeUsername));
      setShowCreate(true);
    } catch (err) {
      setError(err.message);
    }
  }

  function toggleMember(user) {
    setSelectedMembers(prev =>
      prev.some(m => m.id === user.id)
        ? prev.filter(m => m.id !== user.id)
        : [...prev, user]
    );
  }

  function removeMember(userId) {
    setSelectedMembers(prev => prev.filter(m => m.id !== userId));
  }

  function handleCancelCreate() {
    setShowCreate(false);
    setName('');
    setCurrency('GBP');
    setSelectedMembers([]);
    setAvailableUsers([]);
  }

  async function handleCreate(e) {
    e.preventDefault();
    setCreateError('');
    if (selectedMembers.length === 0) {
      setCreateError('Add at least one other member');
      return;
    }
    try {
      await api.createGroup(name, currency, selectedMembers.map(u => u.id));
      handleCancelCreate();
      loadGroups();
    } catch (err) {
      setCreateError(err.message);
    }
  }

  async function handleGenerateInvite() {
    setInviteError('');
    setInviteCode('');
    try {
      const res = await api.generateInvite(inviteIsAdmin);
      setInviteCode(res?.code ?? '');
    } catch (err) {
      setInviteError(err.message);
    }
  }

  function handleCloseInvite() {
    setShowInvite(false);
    setInviteIsAdmin(false);
    setInviteCode('');
    setInviteError('');
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h1>My Groups</h1>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          {isAdmin && <button onClick={() => setShowInvite(v => !v)}>Generate invite</button>}
          <button onClick={onLogout}>Sign out</button>
        </div>
      </div>

      {isAdmin && showInvite && (
        <div className="card">
          <h3>Generate invite</h3>
          <label>
            <input
              type="checkbox"
              checked={inviteIsAdmin}
              onChange={e => setInviteIsAdmin(e.target.checked)}
            />
            {' '}Admin invite
          </label>
          <div style={{ marginTop: '0.5rem' }}>
            <button onClick={handleGenerateInvite}>Generate</button>
            <button type="button" onClick={handleCloseInvite} style={{ marginLeft: '0.5rem' }}>Close</button>
          </div>
          {inviteError && <p className="error">{inviteError}</p>}
          {inviteCode && (
            <div style={{ marginTop: '0.5rem', display: 'flex', gap: '0.5rem' }}>
              <input readOnly value={inviteCode} style={{ flex: 1 }} />
              <button onClick={() => navigator.clipboard.writeText(inviteCode).catch(() => setInviteError('Failed to copy'))}>Copy</button>
            </div>
          )}
        </div>
      )}

      {error && <p className="error">{error}</p>}
      {groups.map(g => (
        <div key={g.id} className="card" style={{ cursor: 'pointer' }} onClick={() => onSelectGroup(g.id)}>
          <strong>{g.name}</strong> <span style={{ color: '#666' }}>({g.currency})</span>
          <div style={{ marginTop: '0.25rem', color: g.netBalance >= 0 ? 'green' : 'red' }}>
            {g.netBalance >= 0 ? `You are owed ${g.currency} ${g.netBalance.toFixed(2)}` : `You owe ${g.currency} ${Math.abs(g.netBalance).toFixed(2)}`}
          </div>
        </div>
      ))}
      {!showCreate
        ? <button onClick={handleShowCreate}>+ New group</button>
        : (
          <div className="card">
            <h3>Create group</h3>
            <form onSubmit={handleCreate}>
              <input placeholder="Group name" value={name} onChange={e => setName(e.target.value)} required />
              <input placeholder="Currency (e.g. EUR)" value={currency} onChange={e => setCurrency(e.target.value)} required />
              <div style={{ border: '1px solid #ccc', borderRadius: '4px', background: '#fff', overflow: 'hidden' }}>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px', padding: '6px', borderBottom: '1px solid #eee', minHeight: '36px', alignItems: 'center' }}>
                  {selectedMembers.map(u => (
                    <span key={u.id} style={{ background: '#3b82f6', color: 'white', borderRadius: '12px', padding: '2px 8px', fontSize: '12px', display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                      {u.id}
                      <button type="button" onClick={() => removeMember(u.id)} style={{ background: 'none', border: 'none', color: 'white', cursor: 'pointer', padding: '0', lineHeight: 1 }}>×</button>
                    </span>
                  ))}
                  {selectedMembers.length === 0 && (
                    <span style={{ color: '#aaa', fontSize: '12px', fontStyle: 'italic' }}>Click below to add members…</span>
                  )}
                </div>
                <div style={{ maxHeight: '140px', overflowY: 'auto' }}>
                  {availableUsers.map(u => {
                    const selected = selectedMembers.some(m => m.id === u.id);
                    return (
                      <div
                        key={u.id}
                        onClick={() => toggleMember(u)}
                        style={{ padding: '6px 10px', cursor: 'pointer', background: selected ? '#eff6ff' : 'transparent', display: 'flex', alignItems: 'center', gap: '8px', fontSize: '13px', borderBottom: '1px solid #f0f0f0' }}
                      >
                        <span style={{ color: selected ? '#3b82f6' : '#ccc', width: '14px' }}>{selected ? '✓' : '○'}</span>
                        {u.id}
                      </div>
                    );
                  })}
                </div>
              </div>
              {createError && <p className="error">{createError}</p>}
              <button type="submit">Create</button>
              <button type="button" onClick={handleCancelCreate} style={{ marginLeft: '0.5rem' }}>Cancel</button>
            </form>
          </div>
        )
      }
    </div>
  );
}
