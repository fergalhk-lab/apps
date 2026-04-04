import { describe, it, expect } from 'vitest'
import { groupInitial } from './sidebar-utils'

describe('groupInitial', () => {
  it('returns the uppercased first character of the name', () => {
    expect(groupInitial('holiday trip')).toBe('H')
  })

  it('works for a single character name', () => {
    expect(groupInitial('A')).toBe('A')
  })

  it('handles leading whitespace', () => {
    expect(groupInitial('  rome')).toBe('R')
  })

  it('returns ? for an empty string', () => {
    expect(groupInitial('')).toBe('?')
  })
})
