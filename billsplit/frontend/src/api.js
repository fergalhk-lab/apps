// frontend/src/api.js
const BASE = '/api';

function getToken() {
  return localStorage.getItem('token');
}

async function request(method, path, body) {
  const headers = { 'Content-Type': 'application/json' };
  const token = getToken();
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 204) return null;
  const data = await res.json();
  if (!res.ok) throw Object.assign(new Error(data.error || 'Request failed'), { status: res.status });
  return data;
}

export const api = {
  register: (username, password, inviteCode) =>
    request('POST', '/auth/register', { username, password, inviteCode }),
  login: (username, password) =>
    request('POST', '/auth/login', { username, password }),

  getGroups: () => request('GET', '/groups'),
  getUsers: () => request('GET', '/users'),
  createGroup: (name, currency, members) =>
    request('POST', '/groups', { name, currency, members }),
  getGroup: (id) => request('GET', `/groups/${id}`),

  getExpenses: (groupId, limit = 20, offset = 0) =>
    request('GET', `/groups/${groupId}/expenses?limit=${limit}&offset=${offset}`),
  addExpense: (groupId, payload) =>
    request('POST', `/groups/${groupId}/expenses`, payload),
  deleteExpense: (groupId, eventId) =>
    request('DELETE', `/groups/${groupId}/expenses/${eventId}`),

  addSettlement: (groupId, payload) =>
    request('POST', `/groups/${groupId}/settlements`, payload),

  leaveGroup: (groupId, username) =>
    request('DELETE', `/groups/${groupId}/members/${username}`),

  generateInvite: (isAdmin) =>
    request('POST', '/admin/invites', { isAdmin }),
};

// NOTE: Only decodes the payload — does NOT verify the signature.
// The server enforces admin authorization; this is UI display only.
export function parseToken(token) {
  try {
    return JSON.parse(atob(token.split('.')[1]));
  } catch {
    return {};
  }
}
