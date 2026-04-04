import { describe, it, expect } from 'vitest'
import { parseToken } from './api'

describe('parseToken', () => {
  it('decodes a valid JWT payload', () => {
    // Build a fake JWT: header.payload.signature
    // payload = { username: 'alice', isAdmin: true }
    const payload = btoa(JSON.stringify({ username: 'alice', isAdmin: true }))
    const token = `header.${payload}.sig`
    const result = parseToken(token)
    expect(result.username).toBe('alice')
    expect(result.isAdmin).toBe(true)
  })

  it('returns safe defaults for a malformed token', () => {
    const result = parseToken('not-a-jwt')
    expect(result.username).toBe('')
    expect(result.isAdmin).toBe(false)
  })
})
