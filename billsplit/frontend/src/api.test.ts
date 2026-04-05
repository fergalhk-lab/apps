// frontend/src/api.test.ts
import { describe, it, expect, beforeEach } from 'vitest'
import { storeIdentity, clearIdentity, getIdentity, USERNAME_KEY, IS_ADMIN_KEY } from './api'

describe('identity localStorage helpers', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('storeIdentity writes username and isAdmin to localStorage', () => {
    storeIdentity('alice', true)
    expect(localStorage.getItem(USERNAME_KEY)).toBe('alice')
    expect(localStorage.getItem(IS_ADMIN_KEY)).toBe('true')
  })

  it('storeIdentity writes isAdmin=false correctly', () => {
    storeIdentity('bob', false)
    expect(localStorage.getItem(IS_ADMIN_KEY)).toBe('false')
  })

  it('clearIdentity removes both keys', () => {
    storeIdentity('alice', true)
    clearIdentity()
    expect(localStorage.getItem(USERNAME_KEY)).toBeNull()
    expect(localStorage.getItem(IS_ADMIN_KEY)).toBeNull()
  })

  it('getIdentity returns stored values', () => {
    storeIdentity('bob', false)
    expect(getIdentity()).toEqual({ username: 'bob', isAdmin: false })
  })

  it('getIdentity returns empty defaults when nothing stored', () => {
    expect(getIdentity()).toEqual({ username: '', isAdmin: false })
  })

  it('getIdentity correctly reads isAdmin=true', () => {
    storeIdentity('admin', true)
    expect(getIdentity().isAdmin).toBe(true)
  })
})
