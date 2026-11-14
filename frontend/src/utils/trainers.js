export const verificationLabels = { pending: 'На проверке', approved: 'Подтверждена', rejected: 'Отклонена' }
export const programLabels = {
  draft: 'Черновик', in_review: 'На модерации', published: 'Опубликована', rejected: 'Отклонена', hidden: 'Скрыта',
}

/** Convert multiline input to unique, trimmed nonempty values. */
export function lines(value) {
  return [...new Set(value.split('\n').map((line) => line.trim()).filter(Boolean))]
}

/** Allow public images only from HTTP(S) or a same-origin path. */
export function publicImage(value) {
  return /^(https?:\/\/|\/(?![\\/]))/i.test(value ?? '') ? value : ''
}

/** Convert a decimal price without binary floating point rounding. */
export function priceMinor(value) {
  const [rubles, kopecks = ''] = String(value).split(/[.,]/)
  return Number(rubles) * 100 + Number(kopecks.padEnd(2, '0'))
}
