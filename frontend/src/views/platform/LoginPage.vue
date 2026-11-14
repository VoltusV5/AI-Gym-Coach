<template>
  <platform-layout
    :title="registering ? 'Создать аккаунт' : 'Вход в Спортик'"
    subtitle="Сохраняйте планы, покупки и результаты в своей учётной записи."
  >
    <form class="platform-panel auth-form" @submit.prevent="submit">
      <label
        >Электронная почта<input
          v-model.trim="email"
          type="email"
          required
          maxlength="200"
          autocomplete="username"
      /></label>
      <label
        >Пароль<input
          v-model="password"
          type="password"
          required
          minlength="8"
          maxlength="72"
          :autocomplete="registering ? 'new-password' : 'current-password'"
      /></label>
      <p v-if="error" role="alert" class="platform-error">{{ error }}</p>
      <div class="platform-actions">
        <button class="primary" :disabled="busy">
          {{ busy ? 'Подождите…' : registering ? 'Зарегистрироваться' : 'Войти' }}</button
        ><button type="button" :disabled="busy" @click="registering = !registering">
          {{ registering ? 'Уже есть аккаунт' : 'Создать аккаунт' }}
        </button>
      </div>
    </form>
  </platform-layout>
</template>
<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { errorMessage } from '@/api/api'
const auth = useAuthStore(),
  route = useRoute(),
  router = useRouter()
const email = ref(''),
  password = ref(''),
  error = ref(''),
  busy = ref(false),
  registering = ref(false)
async function submit() {
  if (busy.value) return
  busy.value = true
  error.value = ''
  try {
    await auth[registering.value ? 'register' : 'login']({
      email: email.value,
      password: password.value,
    })
    password.value = ''
    const target = String(route.query.redirect || '/home')
    await router.replace(target.startsWith('/') && !target.startsWith('//') ? target : '/home')
  } catch (cause) {
    error.value = errorMessage(cause)
  } finally {
    busy.value = false
  }
}
</script>
<style scoped>
.auth-form {
  max-width: 480px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
</style>
