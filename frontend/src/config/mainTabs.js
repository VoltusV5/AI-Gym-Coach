import { barbellOutline, gridOutline, homeOutline, personCircleOutline, statsChartOutline } from 'ionicons/icons'

export const MAIN_TABS = [
  { key: 'main', path: '/home', label: 'Главная', icon: homeOutline },
  { key: 'plans', path: '/plans', label: 'Планы', icon: barbellOutline },
  { key: 'catalog', path: '/marketplace', label: 'Каталог', icon: gridOutline },
  { key: 'progress', path: '/progress', label: 'Прогресс', icon: statsChartOutline },
  { key: 'settings', path: '/settings', label: 'Профиль', icon: personCircleOutline },
]
