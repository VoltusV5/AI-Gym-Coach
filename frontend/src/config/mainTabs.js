/**
 * Порядок слева направо = как в нижнем меню (единый источник для AppTabBar).
 * pathPrefixes: при необходимости можно использовать для сопоставления маршрута с вкладкой.
 */
export const MAIN_TABS = [
  {
    key: 'workout',
    path: '/workout-tools',
    pathPrefixes: ['/workout-tools', '/workout']
  },
  { key: 'notes', path: '/notes', pathPrefixes: ['/notes'] },
  { key: 'main', path: '/home', pathPrefixes: ['/home'] },
  { key: 'nutrition', path: '/nutrition', pathPrefixes: ['/nutrition'] },
  { key: 'settings', path: '/settings', pathPrefixes: ['/settings'] }
]
