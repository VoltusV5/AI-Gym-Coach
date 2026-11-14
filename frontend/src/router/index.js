import { createRouter, createWebHistory } from '@ionic/vue-router'
import { useAuthStore } from '@/stores/auth'
import Welcome from '@/views/Welcome.vue'
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
    path: '/previous-plan',
    name: 'PreviousPlan',
    meta: { session: true },
    component: () => import('@/views/Home.vue'),
  },
  {
    path: '/admin/analytics',
    name: 'AdminAnalytics',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/AnalyticsPage.vue'),
  },
  {
    path: '/trainer/analytics',
    name: 'TrainerAnalytics',
    meta: { roles: ['trainer'] },
    component: () => import('@/views/platform/AnalyticsPage.vue'),
  },
  {
    path: '/admin/exercises',
    name: 'AdminExercises',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/ReferencesPage.vue'),
  },
  {
    path: '/admin/categories',
    name: 'AdminCategories',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/ReferencesPage.vue'),
  },
  {
    path: '/admin/reviews',
    name: 'AdminReviews',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/SocialModerationPage.vue'),
  },
  {
    path: '/admin/reports',
    name: 'AdminReports',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/SocialModerationPage.vue'),
  },
  {
    path: '/recommendations',
    name: 'Recommendations',
    meta: { session: true },
    component: () => import('@/views/platform/RecommendationsPage.vue'),
  },
  {
    path: '/progress',
    name: 'Progress',
    meta: { session: true },
    component: () => import('@/views/platform/ProgressPage.vue'),
  },
  {
    path: '/progress/exercises/:id',
    name: 'ExerciseProgress',
    meta: { session: true },
    component: () => import('@/views/platform/ExerciseProgressPage.vue'),
  },
  {
    path: '/training/:id',
    name: 'TrainingV2',
    meta: { session: true },
    component: () => import('@/views/platform/TrainingPage.vue'),
  },
  {
    path: '/workouts/:id',
    name: 'WorkoutDetail',
    meta: { session: true },
    component: () => import('@/views/platform/WorkoutDetailPage.vue'),
  },
  {
    path: '/saved-shares',
    name: 'SavedShares',
    meta: { registered: true },
    component: () => import('@/views/platform/SavedSharesPage.vue'),
  },
  {
    path: '/saved-shares/:id',
    name: 'SavedShare',
    meta: { registered: true },
    component: () => import('@/views/platform/SharePage.vue'),
  },
  {
    path: '/programs/:id/reviews',
    name: 'ProgramReviews',
    component: () => import('@/views/platform/ReviewsPage.vue'),
  },
  {
    path: '/trainers',
    name: 'MyTrainers',
    meta: { registered: true },
    component: () => import('@/views/platform/MyTrainersPage.vue'),
  },
  {
    path: '/notifications',
    name: 'Notifications',
    meta: { registered: true },
    component: () => import('@/views/platform/NotificationsPage.vue'),
  },
  {
    path: '/share/:token',
    name: 'PublicShare',
    component: () => import('@/views/platform/SharePage.vue'),
  },
  {
    path: '/trainer',
    name: 'TrainerCabinet',
    meta: { roles: ['trainer'] },
    component: () => import('@/views/platform/TrainerPage.vue'),
  },
  {
    path: '/trainers/:id',
    name: 'TrainerPublic',
    component: () => import('@/views/platform/TrainerPage.vue'),
  },
  {
    path: '/author/programs',
    name: 'AuthorPrograms',
    meta: { roles: ['trainer'] },
    component: () => import('@/views/platform/AuthorProgramsPage.vue'),
  },
  {
    path: '/author/programs/new',
    name: 'AuthorProgramNew',
    meta: { roles: ['trainer'] },
    component: () => import('@/views/platform/AuthorEditorPage.vue'),
  },
  {
    path: '/author/programs/:id',
    name: 'AuthorProgramEdit',
    meta: { roles: ['trainer'] },
    component: () => import('@/views/platform/AuthorEditorPage.vue'),
  },
  {
    path: '/trainer/finance',
    name: 'TrainerFinance',
    meta: { roles: ['trainer'] },
    component: () => import('@/views/platform/FinancePage.vue'),
  },
  {
    path: '/admin',
    name: 'Admin',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/AdminPage.vue'),
  },
  {
    path: '/admin/users',
    name: 'AdminUsers',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/AdminUsersPage.vue'),
  },
  {
    path: '/admin/verifications',
    name: 'AdminVerifications',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/ModerationPage.vue'),
  },
  {
    path: '/admin/programs',
    name: 'AdminPrograms',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/ModerationPage.vue'),
  },
  {
    path: '/admin/finance',
    name: 'AdminFinance',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/FinancePage.vue'),
  },
  {
    path: '/admin/orders',
    name: 'AdminOrders',
    meta: { roles: ['admin'] },
    component: () => import('@/views/platform/OrdersPage.vue'),
  },
  { path: '/login', name: 'Login', component: () => import('@/views/platform/LoginPage.vue') },
  {
    path: '/access-denied',
    name: 'AccessDenied',
    component: () => import('@/views/platform/AccessDeniedPage.vue'),
  },
  {
    path: '/catalog',
    name: 'Catalog',
    component: () => import('@/views/platform/CatalogPage.vue'),
  },
  {
    path: '/programs/:id',
    name: 'Program',
    component: () => import('@/views/platform/ProgramPage.vue'),
  },
  {
    path: '/favorites',
    name: 'Favorites',
    meta: { registered: true },
    component: () => import('@/views/platform/FavoritesPage.vue'),
  },
  {
    path: '/profile',
    name: 'ProfileV2',
    meta: { session: true },
    component: () => import('@/views/platform/ProfilePage.vue'),
  },
  {
    path: '/plans',
    name: 'Plans',
    meta: { session: true },
    component: () => import('@/views/platform/PlansPage.vue'),
  },
  {
    path: '/plans/new',
    name: 'NewPlan',
    meta: { session: true },
    component: () => import('@/views/platform/PlanEditorPage.vue'),
  },
  {
    path: '/plans/:id',
    name: 'PlanEditor',
    meta: { session: true },
    component: () => import('@/views/platform/PlanEditorPage.vue'),
  },
  {
    path: '/orders',
    name: 'Orders',
    meta: { registered: true },
    component: () => import('@/views/platform/OrdersPage.vue'),
  },
  {
    path: '/checkout/:id',
    name: 'Checkout',
    meta: { registered: true },
    component: () => import('@/views/platform/CheckoutPage.vue'),
  },
  {
    path: '/',
    name: 'Welcome',
    component: Welcome,
  },
  {
    path: '/body-metrics',
    name: 'BodyMetrics',
    component: BodyMetrics,
  },
  {
    path: '/gender',
    name: 'Gender',
    component: Gender,
  },
  {
    path: '/age',
    name: 'Age',
    component: Age,
  },
  {
    path: '/activity-type',
    name: 'ActivityType',
    component: ActivityType,
  },
  {
    path: '/health-restrictions',
    name: 'HealthRestrictions',
    component: HealthRestrictions,
  },
  {
    path: '/goal-selection',
    name: 'GoalSelection',
    component: GoalSelection,
  },
  {
    path: '/fitness-level',
    name: 'FitnessLevel',
    component: FitnessLevel,
  },
  {
    path: '/training-days',
    name: 'TrainingDays',
    component: TrainingDays,
  },
  {
    path: '/plan-generating',
    name: 'PlanGenerating',
    component: PlanGenerating,
  },
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/platform/DashboardPage.vue'),
  },
  {
    path: '/workout-tools',
    name: 'WorkoutTools',
    component: () => import('@/views/WorkoutToolsPage.vue'),
  },
  {
    path: '/workout/session',
    name: 'WorkoutSession',
    component: () => import('@/views/workout/WorkoutExerciseScreen.vue'),
  },
  {
    path: '/workout/alternatives/:slotIndex',
    name: 'WorkoutAlternatives',
    component: () => import('@/views/workout/WorkoutAlternativesScreen.vue'),
    props: true,
  },
  {
    path: '/workout/success',
    name: 'WorkoutSuccess',
    component: () => import('@/views/workout/WorkoutSuccessScreen.vue'),
  },
  {
    path: '/ai-chat',
    name: 'AiChat',
    component: () => import('@/views/workout/AiChatScreen.vue'),
  },
  {
    path: '/notes',
    name: 'Notes',
    component: () => import('@/views/notes/NotesList.vue'),
  },
  {
    path: '/notes/:id',
    name: 'NoteEditor',
    component: () => import('@/views/notes/NotesEditor.vue'),
    props: true,
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/settings/SettingsPage.vue'),
  },
  {
    path: '/settings/profile',
    name: 'SettingsProfile',
    component: () => import('@/views/settings/SettingsProfilePage.vue'),
  },
  {
    path: '/nutrition',
    name: 'Nutrition',
    component: () => import('@/views/nutrition/NutritionHome.vue'),
  },
  {
    path: '/nutrition/add-meal',
    name: 'NutritionAddMeal',
    component: () => import('@/views/nutrition/NutritionDiaryPage.vue'),
  },
  {
    path: '/nutrition/add-meal/search',
    name: 'NutritionDishSearch',
    component: () => import('@/views/nutrition/NutritionDishSearchPage.vue'),
  },
  {
    path: '/nutrition/add-meal/picker',
    name: 'NutritionMealPicker',
    component: () => import('@/views/nutrition/NutritionMealPickerPage.vue'),
  },
  {
    path: '/nutrition/goals',
    name: 'NutritionGoals',
    component: () => import('@/views/nutrition/NutritionGoalsPage.vue'),
  },
  {
    path: '/nutrition/water',
    name: 'NutritionWater',
    component: () => import('@/views/nutrition/NutritionWaterPage.vue'),
  },
  {
    path: '/nutrition/favorites',
    name: 'NutritionFavorites',
    component: () => import('@/views/nutrition/NutritionFavoritesPage.vue'),
  },
  {
    path: '/nutrition/my-dishes',
    name: 'NutritionMyDishes',
    component: () => import('@/views/nutrition/NutritionMyDishesPage.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.ensureInitialized()
  if ((to.meta.session || to.meta.registered || to.meta.roles) && !auth.token)
    return { path: '/login', query: { redirect: to.fullPath } }
  if (to.meta.registered && !auth.isRegistered)
    return { path: '/login', query: { redirect: to.fullPath } }
  if (to.meta.roles) {
    try {
      await auth.fetchIdentity()
    } catch {
      auth.identity = null
      return '/access-denied'
    }
    if (!to.meta.roles.some((role) => auth.roles.includes(role))) return '/access-denied'
  }
})

export default router
