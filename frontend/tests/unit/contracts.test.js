import { test } from 'node:test'
import assert from 'node:assert/strict'
import { appendPage, localDate } from '../../src/utils/pages.js'
import { assignDay, newPlan, writeDays } from '../../src/utils/plans.js'
import { profilePatch } from '../../src/utils/profile.js'
import { safeReturnPath, sessionClaims } from '../../src/utils/session.js'
import { completion, workoutDraft } from '../../src/utils/workoutV2.js'
import { lines, priceMinor, publicImage } from '../../src/utils/trainers.js'

test('cursor results retain first occurrence and stable order', () => {
  const current = [{ id: '1', name: 'Original' }]
  assert.deepEqual(appendPage(current, [{ id: '1' }, { id: '2' }, { id: '2' }]), [...current, { id: '2' }])
  assert.deepEqual(appendPage(current, null), current)
})

test('profile patch preserves fractional weight, false flags and explicit empty lists', () => {
  const current = { version: 5, weight_kg: 70, age: 28, equipment_ids: ['dumbbell'], injuries_notes: true }
  const draft = { ...current, age: '', weight_kg: 70.25, equipment_ids: [], injuries_notes: false }
  assert.deepEqual(profilePatch(current, draft, Object.keys(current)),
    { version: 5, weight_kg: 70.25, equipment_ids: [], injuries_notes: false })
})

test('sign-in redirects stay on the app and session hints tolerate malformed tokens', () => {
  assert.equal(safeReturnPath('/marketplace?level=beginner'), '/marketplace?level=beginner')
  for (const value of ['//example.com', '/\\example.com', 'https://example.com', '/auth/login?loop=1', '/x\n', ['//example.com']]) {
    assert.equal(safeReturnPath(value), '/settings')
  }
  const claims = { sub: '1', is_anonymous: true }
  assert.deepEqual(sessionClaims(`test.${Buffer.from(JSON.stringify(claims)).toString('base64url')}.test`), claims)
  assert.equal(sessionClaims('malformed'), null)
})

test('plan save uses array positions and leaves optional weight unset', () => {
  const plan = newPlan(' Тестовый план ')
  assert.equal(plan.title, 'Тестовый план')
  assert.match(plan.days[0].id, /^[\da-f-]{36}$/)
  assert.equal(localDate(new Date(2026, 10, 7, 23, 59)), '2026-11-07')
  plan.days[0].exercises = [
    { exercise_id: 8, sets: 3, reps_min: 4, reps_max: 5, rest_sec: 0, weight_kg: '', editorId: 'local' },
    { exercise_id: 9, sets: 3, reps_min: 4, reps_max: 5, rest_sec: 0, weight_kg: 0, editorId: 'local-2' },
  ]
  const [day] = writeDays(plan.days)
  assert.deepEqual(day.exercises[0], { exercise_id: 8, position: 0, sets: 3, reps_min: 4, reps_max: 5, weight_kg: null, rest_sec: 0 })
  assert.equal(day.exercises[1].position, 1)
  const scheduled = assignDay(assignDay([], 5, day.id), 1, day.id)
  assert.deepEqual(scheduled.map(({ weekday }) => weekday), [1, 5])
  assert.deepEqual(assignDay(scheduled, 5, ''), [scheduled[0]])
})

test('workout completion pins the plan revision and sends only actual sets', () => {
  const day = { id: 'day', exercises: [
    { exercise_id: 9, position: 3, weight_kg: 20 },
    { exercise_id: 10, position: 5, weight_kg: null },
  ] }
  const start = new Date('2026-11-07T08:00:00Z')
  const finish = new Date('2026-11-07T08:40:00Z')
  const draft = workoutDraft({ id: 'plan', version: 2 }, day, start)
  draft.slots[0].sets.push({ reps: 8, weight_kg: 21.25, is_warmup: false })
  draft.slots[0].sets.push({ reps: 5, weight_kg: '', is_warmup: true })
  const body = completion(draft, finish)
  assert.equal(body.plan_version, 2)
  assert.equal(body.finished_at, finish.toISOString())
  assert.deepEqual(body.slots, [{ exercise_id: 9, position: 3, sets: [
    { reps: 8, weight_kg: 21.25, is_warmup: false },
    { reps: 5, weight_kg: 0, is_warmup: true },
  ] }])
  assert.equal(draft.slots.length, 2)
})

test('trainer forms normalize references, prices and public image URLs', () => {
  assert.deepEqual(lines(' private/test\n \nprivate/test\nprivate/other '), ['private/test', 'private/other'])
  assert.equal(priceMinor('123,05'), 12305)
  assert.equal(priceMinor('1.1'), 110)
  assert.equal(priceMinor('0'), 0)
  assert.equal(publicImage('javascript:alert(1)'), '')
  assert.equal(publicImage('//example.com/pic'), '')
  assert.equal(publicImage('/static/avatar.png'), '/static/avatar.png')
})
