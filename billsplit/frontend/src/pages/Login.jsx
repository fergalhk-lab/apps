// frontend/src/pages/Login.jsx
import { useState } from 'react';
import { api } from '../api.js';

export default function Login({ onLogin, onRegister }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    try {
      const { token } = await api.login(username, password);
      onLogin(token);
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <div>
      <h1>BillSplit</h1>
      <h2>Sign in</h2>
      <form onSubmit={handleSubmit}>
        <input placeholder="Username" value={username} onChange={e => setUsername(e.target.value)} required />
        <input type="password" placeholder="Password" value={password} onChange={e => setPassword(e.target.value)} required />
        {error && <p className="error">{error}</p>}
        <button type="submit">Sign in</button>
      </form>
      <p style={{marginTop:'1rem'}}>
        No account? <button onClick={onRegister} style={{background:'none',border:'none',color:'blue',padding:0}}>Register</button>
      </p>
    </div>
  );
}
