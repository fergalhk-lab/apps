// frontend/src/App.jsx
import { useState, useEffect, useMemo } from 'react';
import Login from './pages/Login.jsx';
import Register from './pages/Register.jsx';
import GroupList from './pages/GroupList.jsx';
import GroupDetail from './pages/GroupDetail.jsx';
import { parseToken } from './api.js';

export default function App() {
  const [token, setToken] = useState(() => localStorage.getItem('token'));
  const [page, setPage] = useState('login');
  const [selectedGroupId, setSelectedGroupId] = useState(null);

  useEffect(() => {
    if (token) setPage('groups');
    else setPage('login');
  }, [token]);

  function handleLogin(tok) {
    localStorage.setItem('token', tok);
    setToken(tok);
  }

  function handleLogout() {
    localStorage.removeItem('token');
    setToken(null);
  }

  function openGroup(id) {
    setSelectedGroupId(id);
    setPage('group');
  }

  const isAdmin = useMemo(
    () => (token ? parseToken(token).isAdmin === true : false),
    [token]
  );

  if (!token) {
    return page === 'register'
      ? <Register onSuccess={() => setPage('login')} onBack={() => setPage('login')} />
      : <Login onLogin={handleLogin} onRegister={() => setPage('register')} />;
  }

  if (page === 'group' && selectedGroupId) {
    return <GroupDetail groupId={selectedGroupId} onBack={() => setPage('groups')} onLogout={handleLogout} />;
  }

  return <GroupList onSelectGroup={openGroup} onLogout={handleLogout} isAdmin={isAdmin} token={token} />;
}
