// router/index.js
import { createRouter, createWebHistory } from '@ionic/vue-router'
import Welcome from '@/views/Welcome.vue'

// Онбординг: рост/вес → пол → возраст → активность → травмы → цель → уровень → дни
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
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router
