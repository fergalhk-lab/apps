// frontend/src/api.ts

const BASE = '/api'

export const USERNAME_KEY = 'billsplit_username'
export const IS_ADMIN_KEY = 'billsplit_is_admin'

export function storeIdentity(username: string, isAdmin: boolean): void {
  localStorage.setItem(USERNAME_KEY, username)
  localStorage.setItem(IS_ADMIN_KEY, String(isAdmin))
}

export function clearIdentity(): void {
  localStorage.removeItem(USERNAME_KEY)
  localStorage.removeItem(IS_ADMIN_KEY)
}

export function getIdentity(): { username: string; isAdmin: boolean } {
  return {
    username: localStorage.getItem(USERNAME_KEY) ?? '',
    isAdmin: localStorage.getItem(IS_ADMIN_KEY) === 'true',
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 204) return null as T
  const data = await res.json()
  if (!res.ok) {
    if (res.status === 401) {
      clearIdentity()
      window.location.href = '/login'
    }
    throw Object.assign(new Error(data.error || 'Request failed'), { status: res.status })
  }
  return data as T
}

export type Settlement = {
  from: string
  to: string
  amount: number
}

export type Group = {
  id: string
  name: string
  currency: string
  balances: Record<string, number>
  members: string[]
  netBalance: number
  settlements?: Settlement[]
}

export type EventType = 'expense' | 'settlement' | 'reversal'

export type OriginalExpense = {
  currency: string
  amount: number
}

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
  originalExpense?: OriginalExpense
}

export type EventsResponse = {
  events: GroupEvent[]
  total: number
}

export type UserSummary = { id: string }

export type CurrenciesResponse = {
  base: string
  rates: Record<string, number>
  updatedAt: string
}

export const api = {
  register: (username: string, password: string, inviteCode: string) =>
    request<void>('POST', '/auth/register', { username, password, inviteCode }),

  login: async (username: string, password: string): Promise<{ username: string; isAdmin: boolean }> => {
    const res = await request<{ username: string; isAdmin: boolean }>('POST', '/auth/login', { username, password })
    storeIdentity(res.username, res.isAdmin)
    return res
  },

  logout: async (): Promise<void> => {
    await request<void>('POST', '/auth/logout')
    clearIdentity()
  },

  getGroups: () => request<Group[]>('GET', '/groups'),

  getUsers: () => request<{ users: UserSummary[] }>('GET', '/users'),

  createGroup: (name: string, currency: string, members: string[]) =>
    request<{ id: string }>('POST', '/groups', { name, currency, members }),

  getGroup: (id: string) => request<Group>('GET', `/groups/${id}`),

  getExpenses: (groupId: string, limit = 20, offset = 0) =>
    request<EventsResponse>('GET', `/groups/${groupId}/expenses?limit=${limit}&offset=${offset}`),

  addExpense: (groupId: string, payload: {
    description: string
    amount: number
    paidBy: string
    splits: Record<string, number>
    currency?: string
  }) => request<void>('POST', `/groups/${groupId}/expenses`, payload),

  deleteExpense: (groupId: string, eventId: string) =>
    request<void>('DELETE', `/groups/${groupId}/expenses/${eventId}`),

  addSettlement: (groupId: string, payload: { from: string; to: string; amount: number }) =>
    request<void>('POST', `/groups/${groupId}/settlements`, payload),

  leaveGroup: (groupId: string, username: string) =>
    request<void>('DELETE', `/groups/${groupId}/members/${username}`),

  deleteGroup: (id: string) => request<void>('DELETE', `/groups/${id}`),

  generateInvite: (isAdmin: boolean) =>
    request<{ code: string }>('POST', '/admin/invites', { isAdmin }),

  getCurrencies: () => request<CurrenciesResponse>('GET', '/currencies'),
}
