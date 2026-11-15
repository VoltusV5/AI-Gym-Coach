import { expect, it } from 'vitest'
import { metricNumber, metricPeriod, metricRange, weightTrend } from '@/utils/metrics'

it('presents changes without binary tails, negative zero or an implied negative body mass', () => {
  expect(weightTrend(-1.2000000000000028)).toEqual({ value: '1,2 кг', label: 'Вес снизился' })
  expect(weightTrend(0.25)).toEqual({ value: '0,25 кг', label: 'Вес увеличился' })
  expect(weightTrend(-0.000001)).toEqual({ value: '0 кг', label: 'Вес не изменился' })
  expect(weightTrend(-0.005)).toEqual({ value: '0,01 кг', label: 'Вес снизился' })
  expect(weightTrend(null).value).toBe('—')
  expect(metricNumber(Number.NaN)).toBe('—')
})

it('uses the server timezone and exclusive period end for the displayed dates', () => {
  const metrics = {
    period: 'month',
    start: '2026-11-01T00:00:00+03:00',
    end: '2026-12-01T00:00:00+03:00',
    time_zone: 'Europe/Moscow',
  }
  expect(metricPeriod(metrics)).toBe('За месяц')
  expect(metricRange(metrics)).toMatch(/1.*30.*нояб.*2026/)
  expect(metricRange(metrics)).not.toContain('дек')
  expect(metricPeriod({ period: '28d' })).toBe('Последние 28 дней')
  expect(metricRange({})).toBe('')
})
