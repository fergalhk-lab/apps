import { describe, it, expect } from 'vitest'
import { computeSplits } from './AddExpenseModal'

describe('computeSplits', () => {
  const members = ['Alice', 'Bob', 'Carol']

  it('splits equally', () => {
    // 9000 cents = $90.00
    const result = computeSplits('equal', 9000, members, {}, {}, {})
    expect(result).toEqual({ Alice: 3000, Bob: 3000, Carol: 3000 })
  })

  it('splits by shares', () => {
    const shares = { Alice: '2', Bob: '1', Carol: '1' }
    const result = computeSplits('shares', 10000, members, shares, {}, {})
    expect(result?.Alice).toBe(5000)
    expect(result?.Bob).toBe(2500)
    expect(result?.Carol).toBe(2500)
  })

  it('splits by fixed amounts', () => {
    // User enters decimals; computeSplits converts to cents
    const fixed = { Alice: '50', Bob: '30', Carol: '20' }
    const result = computeSplits('fixed', 10000, members, {}, fixed, {})
    expect(result).toEqual({ Alice: 5000, Bob: 3000, Carol: 2000 })
  })

  it('returns null for zero shares sum', () => {
    const shares = { Alice: '0', Bob: '0', Carol: '0' }
    expect(computeSplits('shares', 10000, members, shares, {}, {})).toBeNull()
  })

  it('equal split: sums exactly to total when not evenly divisible', () => {
    // 1000 cents / 3 = 333+333+334 = 1000
    const result = computeSplits('equal', 1000, members, {}, {}, {})!
    const sum = Object.values(result).reduce((a, b) => a + b, 0)
    expect(sum).toBe(1000)
  })

  it('shares: sums exactly to total when rounding would lose a cent', () => {
    const shares = { Alice: '1', Bob: '1', Carol: '1' }
    const result = computeSplits('shares', 1000, members, shares, {}, {})!
    const sum = Object.values(result).reduce((a, b) => a + b, 0)
    expect(sum).toBe(1000)
  })

  it('splits by percentage', () => {
    const pcts = { Alice: '50', Bob: '30', Carol: '20' }
    const result = computeSplits('percentage', 10000, members, {}, {}, pcts)
    expect(result).toEqual({ Alice: 5000, Bob: 3000, Carol: 2000 })
  })

  it('percentage: sums exactly to total when percentages cause rounding', () => {
    const pcts = { Alice: '33.33', Bob: '33.33', Carol: '33.34' }
    const result = computeSplits('percentage', 1000, members, {}, {}, pcts)!
    const sum = Object.values(result).reduce((a, b) => a + b, 0)
    expect(sum).toBe(1000)
  })

  it('percentage: returns null when percentages do not sum to 100', () => {
    const pcts = { Alice: '40', Bob: '40', Carol: '10' }
    expect(computeSplits('percentage', 10000, members, {}, {}, pcts)).toBeNull()
  })

  it('percentage: returns null when total is zero', () => {
    const pcts = { Alice: '33.33', Bob: '33.33', Carol: '33.34' }
    expect(computeSplits('percentage', 0, members, {}, {}, pcts)).toBeNull()
  })
})
