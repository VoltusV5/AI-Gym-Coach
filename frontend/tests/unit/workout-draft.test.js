import { it, expect, beforeEach } from 'vitest'
import { reactive } from 'vue'
import {
  createWorkoutDraft,
  freezeWorkout,
  readWorkoutDraft,
  storeWorkoutDraft,
  removeWorkoutDraft,
} from '@/utils/workoutDraft'
beforeEach(() => sessionStorage.clear())
it('freezes a reactive draft once and restores it only for the same account and plan', () => {
  const plan = { id: 'plan', version: 3 },
    day = {
      id: 'day',
      exercises: [{ exercise_id: 1, position: 0, sets: 2, reps_min: 8, weight_kg: null }],
    }
  const draft = reactive(createWorkoutDraft(plan, day, 101))
  const first = freezeWorkout(draft)
  draft.input.slots[0].sets[0].reps = 12
  expect(freezeWorkout(draft)).toEqual(first)
  expect(first.slots[0].sets[0].reps).toBe(8)
  storeWorkoutDraft(draft)
  expect(readWorkoutDraft(101, 'plan').pending).toEqual(first)
  expect(readWorkoutDraft(102, 'plan')).toBeNull()
  expect(readWorkoutDraft(101, 'another-plan')).toBeNull()
})

it('keeps separate drafts when another plan completes', () => {
  const day = { id: 'day', exercises: [] }
  const a = createWorkoutDraft({ id: 'A', version: 1 }, day, 101),
    b = createWorkoutDraft({ id: 'B', version: 1 }, day, 101)
  storeWorkoutDraft(a)
  storeWorkoutDraft(b)
  removeWorkoutDraft(101, 'A')
  expect(readWorkoutDraft(101, 'B').input.client_session_id).toBe(b.input.client_session_id)
})
