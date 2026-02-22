# 1. Запуск проекта

### Перейди в папку проекта

```sh
cd frontend
```

### Установи зависимости (если ещё не устанавливал)

```sh
npm install
```

### Запусти dev-сервер

```sh
npm run dev
```

### Lint with [ESLint](https://eslint.org/)

```sh
npm run lint
```

### Открой в браузере адрес, который покажет терминал

(обычно http://localhost:5173 или http://localhost:8100).

<br><br><br>

# 2. Создание новых страниц

**Все страницы лежат в папке src/views/**

Пример: создаём страницу «Выбор пола» (Gender.vue).

Перейди в папку src/views/
Создай новый файл → Gender.vue

Вставь минимальный шаблон:

```js
<!-- src/views/Gender.vue -->
<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>Выбор пола</ion-title>
      </ion-toolbar>
    </ion-header>

    <ion-content class="ion-padding">
      <h2>Выберите ваш пол</h2>

      <ion-button expand="block" color="primary" @click="selectGender('male')">
        Мужской
      </ion-button>

      <ion-button expand="block" color="secondary" @click="selectGender('female')">
        Женский
      </ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup>
import { IonPage, IonHeader, IonToolbar, IonTitle, IonContent, IonButton } from '@ionic/vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const selectGender = (gender) => {
  console.log('Выбран пол:', gender)
  router.push('/age')  // переход на следующую страницу
}
</script>

<style scoped>
/* твои стили, если нужно */
ion-button {
  margin: 1rem 0;
}
</style>
```

# 3. Добавь маршрут в роутер

**Открой файл src/router/index.js.**
Добавь импорт и новый маршрут:

```js
// src/router/index.js
import { createRouter, createWebHistory } from '@ionic/vue-router'

// существующие импорты
import Welcome from '@/views/Welcome.vue'
import Gender from '@/views/Gender.vue' // ← новый импорт

const routes = [
  {
    path: '/',
    name: 'Welcome',
    component: Welcome,
  },
  {
    path: '/gender',
    name: 'Gender',
    component: Gender,
  },
  // сюда добавляй следующие страницы позже
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
```

# 4. Сделай переход на новую страницу

Вернись в Welcome.vue и измени функцию перехода:

```js
const startOnboarding = () => {
  router.push('/gender') // ← сюда пишем путь новой страницы
}
```

Или используй <router-link> (если хочешь ссылку вместо кнопки):

```js
<router-link to="/gender">
  <ion-button expand="block" color="primary">
    Начать
  </ion-button>
</router-link>
```

# 5. Чем отличаются папки

**components/ui/**
Самые маленькие переиспользуемые кусочки интерфейса:

- кнопка (BaseButton.vue)
- поле ввода (BaseInput.vue)
- радиокнопки (RadioGroup.vue)
- переключатель Да/Нет (ToggleSwitch.vue)
  Эти компоненты используются на многих страницах, чтобы не писать один и тот же код 10 раз.

**components/layout/**

- Большие «обёртки» или макеты для целых страниц:
- OnboardingLayout.vue — общий контейнер для всех страниц онбординга (с прогресс-баром, отступами, кнопками «Назад/Далее»)
- AppHeader.vue — шапка приложения, если будет
  Это как «рамка», в которую вставляется содержимое страницы.

**views/**
Полноценные экраны приложения:

- Welcome.vue — приветствие
- Gender.vue — выбор пола
- BirthYear.vue — год рождения
  Каждая view — это отдельная страница, которую видит пользователь.
  Внутри view импортируются компоненты из components/ui и components/layout.

### Простая аналогия:

ui — это «кирпичики» (кнопки, инпуты)
layout — это «формы для кирпичиков» (макет страницы)
views — это «готовые комнаты» (страницы), собранные из кирпичиков и форм

# Чек-лист для создания новой страницы

Создай файл Имя.vue в src/views/
Добавь импорт и маршрут в src/router/index.js
Сделай ссылку/переход с предыдущей страницы (router.push('/путь'))
Сохрани → проверь в браузере
