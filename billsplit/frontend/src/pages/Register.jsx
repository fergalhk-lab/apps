// frontend/src/pages/Register.jsx
import { useState } from 'react';
import { api } from '../api.js';

export default function Register({ onSuccess, onBack }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [inviteCode, setInviteCode] = useState('');
  const [error, setError] = useState('');

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    try {
      await api.register(username, password, inviteCode);
      onSuccess();
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <div>
      <h1>BillSplit</h1>
      <h2>Create account</h2>
      <form onSubmit={handleSubmit}>
        <input placeholder="Username" value={username} onChange={e => setUsername(e.target.value)} required />
        <input type="password" placeholder="Password" value={password} onChange={e => setPassword(e.target.value)} required />
        <input placeholder="Invite code" value={inviteCode} onChange={e => setInviteCode(e.target.value)} required />
        {error && <p className="error">{error}</p>}
        <button type="submit">Create account</button>
      </form>
      <p style={{marginTop:'1rem'}}>
        <button onClick={onBack} style={{background:'none',border:'none',color:'blue',padding:0}}>Back to sign in</button>
      </p>
    </div>
  );
}
