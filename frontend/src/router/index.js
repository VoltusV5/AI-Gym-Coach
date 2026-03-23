// router/index.js
import { createRouter, createWebHistory } from '@ionic/vue-router'
import Welcome from '@/views/Welcome.vue'

// Онбординг страницы
import Gender from '@/views/onboarding/Gender.vue'
import BirthYear from '@/views/onboarding/BirthYear.vue'
import BodyMetrics from '@/views/onboarding/BodyMetrics.vue'
import ActivityType from '@/views/onboarding/ActivityType.vue'
import FitnessLevel from '@/views/onboarding/FitnessLevel.vue'
import HealthRestrictions from '@/views/onboarding/HealthRestrictions.vue'
import GoalSelection from '@/views/onboarding/GoalSelection.vue'
import TrainingDays from '@/views/onboarding/TrainingDays.vue'
import PlanGenerating from '@/views/onboarding/PlanGenerating.vue'

const routes = [
  {
    path: '/',
    name: 'Welcome',
    component: Welcome
  },
  {
    path: '/gender',
    name: 'Gender',
    component: Gender
  },
  {
    path: '/birth-year',
    name: 'BirthYear',
    component: BirthYear
  },
  {
    path: '/body-metrics',
    name: 'BodyMetrics',
    component: BodyMetrics
  },
  {
    path: '/activity-type',
    name: 'ActivityType',
    component: ActivityType
  },
  {
    path: '/fitness-level',
    name: 'FitnessLevel',
    component: FitnessLevel
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
