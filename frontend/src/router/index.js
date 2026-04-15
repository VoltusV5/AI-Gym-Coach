// router/index.js
import { createRouter, createWebHistory } from '@ionic/vue-router'
import Welcome from '@/views/Welcome.vue'

// РћРЅР±РѕСЂРґРёРЅРі: СЂРѕСЃС‚/РІРµСЃ в†’ РїРѕР» в†’ РІРѕР·СЂР°СЃС‚ в†’ Р°РєС‚РёРІРЅРѕСЃС‚СЊ в†’ С‚СЂР°РІРјС‹ в†’ С†РµР»СЊ в†’ СѓСЂРѕРІРµРЅСЊ в†’ РґРЅРё
import BodyMetrics from '@/views/onboarding/BodyMetrics.vue'
import Gender from '@/views/onboarding/Gender.vue'
import Age from '@/views/onboarding/Age.vue'
import ActivityType from '@/views/onboarding/ActivityType.vue'
import HealthRestrictions from '@/views/onboarding/HealthRestrictions.vue'
import GoalSelection from '@/views/onboarding/GoalSelection.vue'
import FitnessLevel from '@/views/onboarding/FitnessLevel.vue'
import TrainingDays from '@/views/onboarding/TrainingDays.vue'
import PlanGenerating from '@/views/onboarding/PlanGenerating.vue'

const routes = [
  {
    path: '/',
    name: 'Welcome',
    component: Welcome
  },
  {
    path: '/body-metrics',
    name: 'BodyMetrics',
    component: BodyMetrics
  },
  {
    path: '/gender',
    name: 'Gender',
    component: Gender
  },
  {
    path: '/age',
    name: 'Age',
    component: Age
  },
  {
    path: '/activity-type',
    name: 'ActivityType',
    component: ActivityType
  },
  {
    path: '/health-restrictions',
    name: 'HealthRestrictions',
    component: HealthRestrictions
  },
  {
    path: '/goal-selection',
    name: 'GoalSelection',
    component: GoalSelection
  },
  {
    path: '/fitness-level',
    name: 'FitnessLevel',
    component: FitnessLevel
  },
  {
    path: '/training-days',
    name: 'TrainingDays',
    component: TrainingDays
  },
  {
    path: '/plan-generating',
    name: 'PlanGenerating',
    component: PlanGenerating
  },
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/Home.vue')
  },
  {
    path: '/workout-tools',
    name: 'WorkoutTools',
    component: () => import('@/views/WorkoutToolsPage.vue')
  },
  {
    path: '/workout/session',
    name: 'WorkoutSession',
    component: () => import('@/views/workout/WorkoutExerciseScreen.vue')
  },
  {
    path: '/workout/alternatives/:slotIndex',
    name: 'WorkoutAlternatives',
    component: () => import('@/views/workout/WorkoutAlternativesScreen.vue'),
    props: true
  },
  {
    path: '/notes',
    name: 'Notes',
    component: () => import('@/views/notes/NotesList.vue')
  },
  {
    path: '/notes/:id',
    name: 'NoteEditor',
    component: () => import('@/views/notes/NotesEditor.vue'),
    props: true
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/settings/SettingsPage.vue')
  },
  {
    path: '/settings/profile',
    name: 'SettingsProfile',
    component: () => import('@/views/settings/SettingsProfilePage.vue')
  },
  {
    path: '/nutrition',
    name: 'Nutrition',
    component: () => import('@/views/NutritionPlaceholder.vue')
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router

