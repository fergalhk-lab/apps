// frontend-new/src/api.ts

const BASE = '/api'
export const TOKEN_KEY = 'billsplit_token'

function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 204) return null as T
  const data = await res.json()
  if (!res.ok) throw Object.assign(new Error(data.error || 'Request failed'), { status: res.status })
  return data as T
}

export type Group = {
  id: string
  name: string
  currency: string
  balances: Record<string, number>
  members: string[]
  netBalance: number
}

export type EventType = 'expense' | 'settlement' | 'reversal'

export type GroupEvent = {
  id: string
  type: EventType
  description?: string
  amount: number
  paidBy?: string
  from?: string
  to?: string
  splits?: Record<string, number>
  createdAt: string
  reversedEventId?: string
}

export type EventsResponse = {
  events: GroupEvent[]
  total: number
}

export type UserSummary = { id: string }

export type TokenPayload = { username: string; isAdmin: boolean }

export const api = {
  register: (username: string, password: string, inviteCode: string) =>
    request<void>('POST', '/auth/register', { username, password, inviteCode }),

  login: (username: string, password: string) =>
    request<{ token: string }>('POST', '/auth/login', { username, password }),

  getGroups: () => request<Group[]>('GET', '/groups'),

  getUsers: () => request<{ users: UserSummary[] }>('GET', '/users'),

  createGroup: (name: string, currency: string, members: string[]) =>
    request<Group>('POST', '/groups', { name, currency, members }),

  getGroup: (id: string) => request<Group>('GET', `/groups/${id}`),

  getExpenses: (groupId: string, limit = 20, offset = 0) =>
    request<EventsResponse>('GET', `/groups/${groupId}/expenses?limit=${limit}&offset=${offset}`),

  addExpense: (groupId: string, payload: {
    description: string
    amount: number
    paidBy: string
    splits: Record<string, number>
  }) => request<void>('POST', `/groups/${groupId}/expenses`, payload),

  deleteExpense: (groupId: string, eventId: string) =>
    request<void>('DELETE', `/groups/${groupId}/expenses/${eventId}`),

  addSettlement: (groupId: string, payload: { from: string; to: string; amount: number }) =>
    request<void>('POST', `/groups/${groupId}/settlements`, payload),

  leaveGroup: (groupId: string, username: string) =>
    request<void>('DELETE', `/groups/${groupId}/members/${username}`),

  generateInvite: (isAdmin: boolean) =>
    request<{ code: string }>('POST', '/admin/invites', { isAdmin }),
}

// NOTE: Only decodes the payload — does NOT verify the signature.
// The server enforces admin authorisation; this is UI display only.
export function parseToken(token: string): TokenPayload {
  try {
    const b64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded = b64 + '='.repeat((4 - (b64.length % 4)) % 4)
    return JSON.parse(atob(padded)) as TokenPayload
  } catch {
    return { username: '', isAdmin: false }
  }
}
