<script setup>
import { computed, ref } from 'vue'
import RequestState from '@/components/ui/RequestState.vue'
import { apiStatus, errorText } from '@/api/v2'
import { useWrite } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const emit = defineEmits(['signed-in'])
const modes = computed(() => [
  { id: 'login', label: 'Вход' },
  ...(!auth.token || auth.isGuest ? [{ id: 'register', label: 'Регистрация' }] : []),
  ...(auth.token && !auth.isGuest ? [{ id: 'password', label: 'Смена пароля' }] : []),
])
const mode = ref('login')
const email = ref('')
const password = ref('')
const newPassword = ref('')
const changed = ref(false)
const { saving, error, run } = useWrite()

/** Switch forms without carrying a password into a different operation. */
function select(next) {
  mode.value = next
  password.value = ''
  newPassword.value = ''
  error.value = ''
  changed.value = false
}

/** Submit one authenticated operation and leave input intact on failure. */
function submit() {
  return run(async () => {
    if (mode.value === 'password') {
      await auth.changePassword({ currentPassword: password.value, newPassword: newPassword.value })
      password.value = ''
      newPassword.value = ''
      changed.value = true
      return
    }
    const credentials = { email: email.value, password: password.value }
    if (mode.value === 'register') await auth.register(credentials)
    else await auth.login(credentials)
    password.value = ''
    emit('signed-in')
  }, (err) => {
    if (apiStatus(err) === 401 && mode.value === 'register') {
      return 'Гостевая сессия истекла. Повторите отправку, чтобы создать аккаунт.'
    }
    if (apiStatus(err) === 401) return 'Проверьте email и пароль.'
    if (apiStatus(err) === 409 && mode.value === 'register') return 'Аккаунт с таким email уже существует.'
    return errorText(err)
  })
}
</script>

<template>
  <section class="feature-card" aria-label="Управление аккаунтом">
    <div class="feature-row">
      <button v-for="item in modes" :key="item.id" class="feature-button" type="button"
        :class="{ secondary: item.id !== mode }" :aria-pressed="item.id === mode" :disabled="saving" @click="select(item.id)">
        {{ item.label }}
      </button>
    </div>
    <form class="feature-stack" @submit.prevent="submit">
      <fieldset class="feature-stack" :disabled="saving">
        <legend>{{ modes.find(({ id }) => id === mode)?.label }}</legend>
        <label v-if="mode !== 'password'">Email <input v-model.trim="email" required type="email" autocomplete="email" /></label>
        <label>{{ mode === 'password' ? 'Текущий пароль' : 'Пароль' }}
          <input v-model="password" type="password" :autocomplete="mode === 'register' ? 'new-password' : 'current-password'"
            required :minlength="mode === 'register' ? 8 : undefined" />
        </label>
        <label v-if="mode === 'password'">Новый пароль <input v-model="newPassword" required type="password" minlength="8" autocomplete="new-password" /></label>
      </fieldset>
      <button class="feature-button" :disabled="saving">
        {{ saving ? 'Отправляем…' : mode === 'login' ? 'Войти' : mode === 'register' ? 'Создать аккаунт' : 'Сменить пароль' }}
      </button>
      <RequestState :error="error" />
      <p v-if="changed" class="feature-success" role="status">Пароль обновлён</p>
    </form>
  </section>
</template>
