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

  it('decodes a base64url-encoded JWT payload (real JWT format)', () => {
    // Real JWTs use base64url: '+' → '-', '/' → '_', no padding
    const json = JSON.stringify({ username: 'bob', isAdmin: false })
    const b64url = btoa(json).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
    const token = `header.${b64url}.sig`
    const result = parseToken(token)
    expect(result.username).toBe('bob')
    expect(result.isAdmin).toBe(false)
  })

  it('returns safe defaults for a malformed token', () => {
    const result = parseToken('not-a-jwt')
    expect(result.username).toBe('')
    expect(result.isAdmin).toBe(false)
  })
})
