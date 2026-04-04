// frontend-new/src/components/expense.test.ts
import { describe, it, expect } from 'vitest'
import { computeSplits } from './AddExpenseModal'

describe('computeSplits', () => {
  const members = ['Alice', 'Bob', 'Carol']

  it('splits equally', () => {
    const result = computeSplits('equal', 90, members, {}, {})
    expect(result).toEqual({ Alice: 30, Bob: 30, Carol: 30 })
  })

  it('splits by ratio', () => {
    const ratios = { Alice: '2', Bob: '1', Carol: '1' }
    const result = computeSplits('ratio', 100, members, ratios, {})
    expect(result?.Alice).toBe(50)
    expect(result?.Bob).toBe(25)
    expect(result?.Carol).toBe(25)
  })

  it('splits by fixed amounts', () => {
    const fixed = { Alice: '50', Bob: '30', Carol: '20' }
    const result = computeSplits('fixed', 100, members, {}, fixed)
    expect(result).toEqual({ Alice: 50, Bob: 30, Carol: 20 })
  })

  it('returns null for zero ratio sum', () => {
    const ratios = { Alice: '0', Bob: '0', Carol: '0' }
    expect(computeSplits('ratio', 100, members, ratios, {})).toBeNull()
  })
})
