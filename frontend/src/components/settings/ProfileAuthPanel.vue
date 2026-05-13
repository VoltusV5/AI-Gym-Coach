<template>
  <div class="profile-panel">
    <div class="profile-head">
      <div class="avatar">{{ t.avatar }}</div>
      <div>
        <p class="profile-title">{{ t.title }}</p>
        <p class="profile-sub">{{ t.subtitle }}</p>
      </div>
    </div>

    <div class="seg-row">
      <button
        v-for="mode in modes"
        :key="mode.id"
        type="button"
        class="seg-btn"
        :class="{ 'seg-btn--active': activeMode === mode.id }"
        @click="activeMode = mode.id"
      >
        {{ mode.label }}
      </button>
    </div>

    <div v-if="activeMode === 'login'" class="form-grid">
      <ion-input v-model="email" class="field" type="email" placeholder="Email" />
      <ion-input v-model="password" class="field" type="password" :placeholder="t.password" />
      <ion-button class="sportik-footer-btn" expand="block" :disabled="loading" @click="onLogin">
        {{ t.login }}
      </ion-button>
    </div>

    <div v-else-if="activeMode === 'register'" class="form-grid">
      <ion-input v-model="name" class="field" :placeholder="t.name" />
      <ion-input v-model="email" class="field" type="email" placeholder="Email" />
      <ion-input v-model="password" class="field" type="password" :placeholder="t.password" />
      <ion-button class="sportik-footer-btn" expand="block" :disabled="loading" @click="onRegister">
        {{ t.registerAction }}
      </ion-button>
    </div>

    <div v-else class="form-grid">
      <ion-input v-model="email" class="field" type="email" placeholder="Email" />
      <ion-input v-model="password" class="field" type="password" :placeholder="t.currentPassword" />
      <ion-input v-model="newPassword" class="field" type="password" :placeholder="t.newPassword" />
      <ion-button class="sportik-footer-btn" expand="block" :disabled="loading" @click="onChangePassword">
        {{ t.changePasswordAction }}
      </ion-button>
    </div>

  </div>
</template>

<script setup>
import { ref } from 'vue'
import { IonButton, IonInput, toastController } from '@ionic/vue'
import { useAuthStore } from '@/stores/auth'

const t = {
  avatar: '\u041f',
  title: '\u041f\u0440\u043e\u0444\u0438\u043b\u044c',
  subtitle: '\u0420\u0435\u0433\u0438\u0441\u0442\u0440\u0430\u0446\u0438\u044f / \u0432\u0445\u043e\u0434 / \u0441\u043c\u0435\u043d\u0430 \u043f\u0430\u0440\u043e\u043b\u044f',
  login: '\u0412\u0445\u043e\u0434',
  register: '\u0420\u0435\u0433\u0438\u0441\u0442\u0440\u0430\u0446\u0438\u044f',
  changePassword: '\u0421\u043c\u0435\u043d\u0430 \u043f\u0430\u0440\u043e\u043b\u044f',
  registerAction: '\u0417\u0430\u0440\u0435\u0433\u0438\u0441\u0442\u0440\u0438\u0440\u043e\u0432\u0430\u0442\u044c\u0441\u044f',
  changePasswordAction: '\u0421\u043c\u0435\u043d\u0438\u0442\u044c \u043f\u0430\u0440\u043e\u043b\u044c',
  password: '\u041f\u0430\u0440\u043e\u043b\u044c',
  currentPassword: '\u0422\u0435\u043a\u0443\u0449\u0438\u0439 \u043f\u0430\u0440\u043e\u043b\u044c',
  newPassword: '\u041d\u043e\u0432\u044b\u0439 \u043f\u0430\u0440\u043e\u043b\u044c',
  name: '\u0418\u043c\u044f',
}

const modes = [
  { id: 'login', label: t.login },
  { id: 'register', label: t.register },
  { id: 'change_password', label: t.changePassword }
]

const activeMode = ref('login')
const name = ref('')
const email = ref('')
const password = ref('')
const newPassword = ref('')
const loading = ref(false)
const auth = useAuthStore()

async function showToast(message, color = 'medium') {
  const toast = await toastController.create({
    message,
    duration: 2000,
    color
  })
  await toast.present()
}

function validateEmail(value) {
  return /.+@.+\..+/.test(String(value ?? '').trim())
}

async function onLogin() {
  if (loading.value) return
  if (!validateEmail(email.value) || !password.value) {
    await showToast('Введите корректный email и пароль', 'warning')
    return
  }
  loading.value = true
  try {
    await auth.login({ email: email.value.trim(), password: password.value })
    await auth.fetchProfile()
    await showToast('Вход выполнен', 'success')
  } catch (e) {
    await showToast(e?.response?.data?.message || 'Ошибка входа', 'danger')
  } finally {
    loading.value = false
  }
}

async function onRegister() {
  if (loading.value) return
  if (!validateEmail(email.value) || !password.value || password.value.length < 6) {
    await showToast('Проверьте email и пароль (минимум 6 символов)', 'warning')
    return
  }
  loading.value = true
  try {
    await auth.register({
      email: email.value.trim(),
      password: password.value,
      name: name.value.trim()
    })
    await auth.fetchProfile()
    await showToast('Регистрация выполнена', 'success')
  } catch (e) {
    await showToast(e?.response?.data?.message || 'Ошибка регистрации', 'danger')
  } finally {
    loading.value = false
  }
}

async function onChangePassword() {
  if (loading.value) return
  if (!password.value || !newPassword.value || newPassword.value.length < 6) {
    await showToast('Заполните текущий и новый пароль (минимум 6 символов)', 'warning')
    return
  }
  loading.value = true
  try {
    await auth.changePassword({
      currentPassword: password.value,
      newPassword: newPassword.value
    })
    password.value = ''
    newPassword.value = ''
    await showToast('Пароль обновлён', 'success')
  } catch (e) {
    await showToast(e?.response?.data?.message || 'Ошибка смены пароля', 'danger')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.profile-panel { background: var(--sportik-surface); border: 1px solid var(--sportik-border); border-radius: 16px; padding: 12px; color: var(--sportik-text); box-shadow: var(--sportik-shadow-md); }
.profile-head { display: flex; align-items: center; gap: 10px; }
.avatar { width: 44px; height: 44px; border-radius: 50%; background: var(--sportik-surface-soft); border: 1px solid var(--sportik-border); display: flex; align-items: center; justify-content: center; font-size: 1.1rem; font-weight: 700; }
.profile-title { margin: 0; font-weight: 700; }
.profile-sub { margin: 2px 0 0; font-size: 0.8rem; color: var(--sportik-text-muted); }
.seg-row { display: flex; gap: 6px; margin: 12px 0 10px; }
.seg-btn { border: 1px solid var(--sportik-border); border-radius: 999px; padding: 8px 10px; font-size: 0.76rem; background: var(--sportik-surface-soft); cursor: pointer; color: var(--sportik-text); }
.seg-btn--active { background: color-mix(in srgb, var(--sportik-brand) 25%, var(--sportik-surface)); border-color: color-mix(in srgb, var(--sportik-brand) 60%, transparent); }
.form-grid { display: flex; flex-direction: column; gap: 8px; }
.field { --background: var(--sportik-surface-soft); --color: var(--sportik-text); border-radius: 10px; border: 1px solid var(--sportik-border); }
</style>