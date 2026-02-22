// Pinia store для хранения токена и профиля

import { defineStore } from 'pinia'
// import api from '@/api/api'
// import { saveToken, getToken, removeToken } from '@/utils/authStorage'

/**
 * ТЗ: Реализация Pinia-стора 'auth'
 * 1. State: хранение токена, ID, флага анонимности и данных профиля.
 * 2. Метод init(): 
 *    - Загрузка токена;
 *    - При успехе — получение профиля (GET /profile);
 *    - При отсутствии токена или 401 ошибке — вызов guestLogin.
 * 3. Метод guestLogin(): 
 *    - Регистрация гостя (POST /auth/guest);
 *    - Сохранение токена и данных в state.
 */


// Пример:

// export const useAuthStore = defineStore('auth', {
//   state: () => ({
//     // token: null,
//     // userId: null,
//     // isAnonymous: true,
//     // profile: {}
//   }),

//   actions: {
//     // async init() {
//     //   // this.token = await getToken()
//     //   // ... вся логика гостевого входа и GET /profile
//     // },

//     // async guestLogin() {
//     //   // ...
//     // }
//   }
// })