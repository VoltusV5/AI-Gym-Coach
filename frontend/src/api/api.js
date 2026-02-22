// здесь будет axios с базовым URL и перехватчиком ошибок

import axios from 'axios'

const api = axios.create({
  baseURL: 'http://localhost:9091', // бэкенд
  timeout: 10000,
})

// Заполнение токена из auth store
// ...

// Обработка 401 (токен истёк)
// ...

export default api