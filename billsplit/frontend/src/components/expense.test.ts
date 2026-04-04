import { describe, it, expect } from 'vitest'
import { computeSplits } from './AddExpenseModal'

describe('computeSplits', () => {
  const members = ['Alice', 'Bob', 'Carol']

  it('splits equally', () => {
    const result = computeSplits('equal', 90, members, {}, {}, {})
    expect(result).toEqual({ Alice: 30, Bob: 30, Carol: 30 })
  })

  it('splits by shares', () => {
    const shares = { Alice: '2', Bob: '1', Carol: '1' }
    const result = computeSplits('shares', 100, members, shares, {}, {})
    expect(result?.Alice).toBe(50)
    expect(result?.Bob).toBe(25)
    expect(result?.Carol).toBe(25)
  })

  it('splits by fixed amounts', () => {
    const fixed = { Alice: '50', Bob: '30', Carol: '20' }
    const result = computeSplits('fixed', 100, members, {}, fixed, {})
    expect(result).toEqual({ Alice: 50, Bob: 30, Carol: 20 })
  })

  it('returns null for zero shares sum', () => {
    const shares = { Alice: '0', Bob: '0', Carol: '0' }
    expect(computeSplits('shares', 100, members, shares, {}, {})).toBeNull()
  })

  it('equal split: sums exactly to total when not evenly divisible', () => {
    // $10 / 3 = $3.33 * 3 = $9.99 without largest remainder
    const result = computeSplits('equal', 10, members, {}, {}, {})!
    const sum = Object.values(result).reduce((a, b) => a + b, 0)
    expect(parseFloat(sum.toFixed(2))).toBe(10)
  })

  it('shares: sums exactly to total when rounding would lose a cent', () => {
    // Alice:1, Bob:1, Carol:1 of $10 — each gets $3.33 without fix, $9.99 total
    const shares = { Alice: '1', Bob: '1', Carol: '1' }
    const result = computeSplits('shares', 10, members, shares, {}, {})!
    const sum = Object.values(result).reduce((a, b) => a + b, 0)
    expect(parseFloat(sum.toFixed(2))).toBe(10)
  })

  it('splits by percentage', () => {
    const pcts = { Alice: '50', Bob: '30', Carol: '20' }
    const result = computeSplits('percentage', 100, members, {}, {}, pcts)
    expect(result).toEqual({ Alice: 50, Bob: 30, Carol: 20 })
  })

  it('percentage: sums exactly to total when percentages cause rounding', () => {
    // 33.33% each — would be $3.333 each without largest remainder
    const pcts = { Alice: '33.33', Bob: '33.33', Carol: '33.34' }
    const result = computeSplits('percentage', 10, members, {}, {}, pcts)!
    const sum = Object.values(result).reduce((a, b) => a + b, 0)
    expect(parseFloat(sum.toFixed(2))).toBe(10)
  })

  it('percentage: returns null when percentages do not sum to 100', () => {
    const pcts = { Alice: '40', Bob: '40', Carol: '10' }
    expect(computeSplits('percentage', 100, members, {}, {}, pcts)).toBeNull()
  })

  it('percentage: returns null when total is zero', () => {
    const pcts = { Alice: '33.33', Bob: '33.33', Carol: '33.34' }
    expect(computeSplits('percentage', 0, members, {}, {}, pcts)).toBeNull()
  })
})
