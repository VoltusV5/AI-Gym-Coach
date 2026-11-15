<template>
  <platform-layout
    :title="own ? 'Кабинет тренера' : profile?.display_name || 'Профиль тренера'"
    subtitle="Профессиональные данные и опубликованные программы."
  >
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="fetchProfile">
      <section v-if="profile" class="platform-panel">
        <span v-if="profile.verified" class="platform-badge">Проверенный тренер</span>
        <p v-else class="platform-muted">Профессиональный статус пока не подтверждён</p>
        <img v-if="photoURL" :src="photoURL" :alt="profile.display_name" class="trainer-photo" />
        <h2>{{ profile.display_name }}</h2>
        <p>{{ profile.bio }}</p>
        <div v-if="!own && auth.isRegistered" class="platform-actions">
          <button :disabled="action.busy.value" @click="follow">
            {{ profile.following ? 'Отписаться' : 'Подписаться' }}</button
          ><report-form target-type="trainer" :target-id="profile.id" />
        </div>
        <p>{{ profile.specializations.join(', ') }}</p>
        <p>Опыт: {{ profile.experience_years }} лет</p>
        <p>{{ profile.qualifications }}</p>
        <router-link :to="{ path: '/catalog', query: { trainer_id: profile.id } }"
          >Программы тренера</router-link
        >
        <div v-if="own" class="platform-actions">
          <router-link to="/author/programs">Мои программы</router-link
          ><router-link to="/trainer/finance">Финансы</router-link
          ><router-link to="/trainer/analytics">Статистика</router-link>
        </div>
      </section>
      <form v-if="own" class="platform-panel" @submit.prevent="save">
        <fieldset :disabled="action.busy.value" class="trainer-fields">
          <h2>Профессиональный профиль</h2>
          <label
            >Фото, URL или путь к файлу<input v-model="form.photo_ref" maxlength="2000"
          /></label>
          <div class="platform-fields">
            <label
              >Имя для публикаций<input
                v-model.trim="form.display_name"
                required
                maxlength="200" /></label
            ><label
              >Опыт, лет<input
                v-model.number="form.experience_years"
                type="number"
                min="0"
                max="100"
                required /></label
            ><label>Специализации через запятую<input v-model="specializations" /></label>
          </div>
          <label>О себе<textarea v-model="form.bio" maxlength="10000" /></label
          ><label
            >Квалификация для публичного профиля<textarea
              v-model="form.qualifications"
              maxlength="10000"
            />
          </label>
          <div class="platform-actions">
            <button class="primary" :disabled="action.busy.value">Сохранить профиль</button>
          </div>
        </fieldset>
      </form>
      <section v-if="own" class="platform-panel">
        <h2>Проверка квалификации</h2>
        <p v-if="verification">
          Статус:
          {{
            verification.status === 'pending'
              ? 'На проверке'
              : verification.status === 'approved'
                ? 'Подтверждено'
                : 'Нужны уточнения'
          }}
        </p>
        <p v-if="verification?.reason">{{ verification.reason }}</p>
        <p v-if="verificationError" role="alert" class="platform-error">{{ verificationError }}</p>
        <form
          v-if="!profile?.verified && verification?.status !== 'pending'"
          @submit.prevent="submitVerification"
        >
          <label
            >Сведения о квалификации<textarea
              v-model="qualification"
              required
              maxlength="10000"
            /></label
          ><label
            >Ссылки на закрытые подтверждающие материалы, по одной на строку<textarea
              v-model="documents"
              required
            />
          </label>
          <p class="platform-muted">
            Эти материалы доступны только вам и администрации. Укажите ссылки на уже подготовленные
            подтверждения.
          </p>
          <button class="primary" :disabled="action.busy.value">Отправить на проверку</button>
        </form>
      </section>
      <p v-if="action.error.value" role="alert" class="platform-error">{{ action.error.value }}</p>
      <p v-if="action.success.value" role="status" class="platform-success">
        {{ action.success.value }}
      </p>
    </async-state>
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import ReportForm from '@/components/platform/ReportForm.vue'
import { useRoute } from 'vue-router'
import api, { errorMessage } from '@/api/api'
import { useRequest, useAction } from '@/composables/useAsync'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const auth = useAuthStore(),
  route = useRoute(),
  own = computed(() => !route.params.id),
  load = useRequest(),
  action = useAction()
const photoURL = computed(() => {
  const value = load.data.value?.photo_url || ''
  return /^(https?:\/\/|\/(?!\/))/.test(value) ? value : ''
})
const profile = computed(() => load.data.value),
  form = ref({ display_name: '', bio: '', experience_years: 0, qualifications: '' }),
  specializations = ref(''),
  qualification = ref(''),
  documents = ref(''),
  verification = ref(null),
  verificationError = ref('')
async function fetchProfile() {
  const value = await load.run(async () => {
    try {
      return (
        await api.get(own.value ? '/api/v2/trainer/profile' : `/api/v2/trainers/${route.params.id}`)
      ).data
    } catch (cause) {
      if (own.value && cause.response?.status === 404) return null
      throw cause
    }
  })
  if (value) {
    form.value = {
      display_name: value.display_name,
      photo_ref: value.photo_url || '',
      bio: value.bio,
      experience_years: value.experience_years,
      qualifications: value.qualifications,
    }
    specializations.value = value.specializations.join(', ')
  }
  if (own.value) await fetchVerification()
}
async function fetchVerification() {
  verificationError.value = ''
  try {
    verification.value = (await api.get('/api/v2/trainer/verification')).data
  } catch (cause) {
    if (cause.response?.status !== 404) verificationError.value = errorMessage(cause)
  }
}
async function follow() {
  if (
    await action.run(`/api/v2/trainers/${profile.value.id}/follow`, undefined, {
      method: profile.value.following ? 'delete' : 'put',
      message: 'Подписка обновлена',
    })
  )
    await fetchProfile()
}
async function save() {
  if (
    await action.run(
      '/api/v2/trainer/profile',
      {
        ...form.value,
        photo_ref: form.value.photo_ref || null,
        specializations: specializations.value
          .split(',')
          .map((x) => x.trim())
          .filter(Boolean),
      },
      { method: 'patch' },
    )
  )
    await fetchProfile()
}
async function submitVerification() {
  const result = await action.run(
    '/api/v2/trainer/verification',
    {
      qualification: qualification.value,
      document_refs: documents.value
        .split('\n')
        .map((x) => x.trim())
        .filter(Boolean),
    },
    { message: 'Заявка отправлена' },
  )
  if (result) {
    verification.value = result
    documents.value = ''
    qualification.value = ''
  }
}
useViewLoad(fetchProfile)
</script>

<style scoped>
.trainer-photo {
  width: 110px;
  height: 110px;
  object-fit: cover;
  border-radius: 24px;
}
.trainer-fields {
  border: 0;
  padding: 0;
  margin: 0;
  min-width: 0;
}
</style>
