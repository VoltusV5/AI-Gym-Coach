import { numberLabel } from './workoutV2.js'

export function metricNumber(value) {
  return typeof value === 'number' && Number.isFinite(value)
    ? numberLabel.format(value === 0 ? 0 : value)
    : '—'
}

export function weightTrend(value) {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    return { value: '—', label: 'Недостаточно измерений' }
  }
  const amount = metricNumber(Math.abs(value))
  const rounded = amount === metricNumber(0) ? 0 : value
  return {
    value: `${amount} кг`,
    label: rounded < 0 ? 'Вес снизился' : rounded > 0 ? 'Вес увеличился' : 'Вес не изменился',
  }
}

export function metricPeriod(metrics = {}) {
  return (
    { '28d': 'Последние 28 дней', week: 'За неделю', month: 'За месяц', all: 'За всё время' }[
      metrics.period
    ] || 'За выбранный период'
  )
}

export function metricRange(metrics = {}) {
  const start = new Date(metrics.start)
  const end = new Date(new Date(metrics.end).getTime() - 1)
  if (!Number.isFinite(start.getTime()) || !Number.isFinite(end.getTime()) || end < start) return ''
  try {
    return new Intl.DateTimeFormat('ru-RU', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      timeZone: metrics.time_zone || 'UTC',
    }).formatRange(start, end)
  } catch {
    return ''
  }
}
