<script setup>
import { computed, ref, watch } from 'vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { apiStatus, v2 } from '@/api/v2'
import { useRead, useWrite } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'
import { stashDraft, takeDraft } from '@/utils/reauthDrafts'
import { lines, publicImage } from '@/utils/trainers'

const props = defineProps({ trainerId: { type: String, default: '' } })
const auth = useAuthStore()
const isPublic = computed(() => Boolean(props.trainerId))
const draft = ref(null)
const saved = ref(false)
const { data, loading, error, status, load } = useRead(async (signal) => {
  try {
    const path = isPublic.value ? `/trainers/${encodeURIComponent(props.trainerId)}` : '/trainer/profile'
    return await v2('GET', path, { signal, anonymous: isPublic.value })
  } catch (err) {
    if (isPublic.value || apiStatus(err) !== 404) throw err
    return { display_name: '', photo_url: null, bio: '', specializations: [], experience_years: 0, qualifications: '', verified: false }
  }
}, 'Публичный профиль тренера не найден.')
const { saving, error: saveError, status: saveStatus, run } = useWrite()
watch([() => props.trainerId, () => !isPublic.value && auth.token], () => {
  if (!isPublic.value && !auth.token && draft.value) stashDraft(auth.lastSubject, 'trainer-profile', draft.value)
  draft.value = null
  load()
}, { immediate: true })

/** Prepare the form from the server profile, keeping the verified flag read-only. */
function setProfile(profile) {
  draft.value = {
    display_name: profile.display_name, photo_ref: profile.photo_url ?? '', bio: profile.bio,
    specializations: (profile.specializations ?? []).join('\n'),
    experience_years: profile.experience_years, qualifications: profile.qualifications,
  }
}
watch(data, (profile) => {
  if (profile && !isPublic.value) {
    const previous = takeDraft(auth.accountId, 'trainer-profile')
    if (previous) draft.value = previous
    else setProfile(profile)
  }
})

/** Update professional profile fields, excluding verification and public review data. */
function save() {
  return run(async () => {
    saved.value = false
    data.value = await v2('PATCH', '/trainer/profile', {
      data: { ...draft.value, photo_ref: draft.value.photo_ref || null, specializations: lines(draft.value.specializations) },
    })
    saved.value = true
  })
}
</script>

<template>
  <FeaturePage :eyebrow="isPublic ? 'Тренеры' : 'Кабинет тренера'" :title="isPublic ? data?.display_name ?? 'Профиль тренера' : 'Профессиональный профиль'"
    :back="isPublic ? '/marketplace' : '/settings'">
    <div class="feature-stack">
      <RequestState :loading="loading" :error="error" :status="status" :retry="load" />
      <section v-if="isPublic && data" class="feature-card">
        <img v-if="publicImage(data.photo_url)" class="feature-avatar" :src="publicImage(data.photo_url)" alt="" />
        <span v-if="data.verified" class="feature-badge">Тренер подтверждён</span>
        <p class="feature-text">{{ data.bio || 'Тренер пока не добавил описание.' }}</p>
        <p v-if="data.specializations?.length">Специализации: {{ data.specializations.join(', ') }}</p>
        <p>Опыт: {{ data.experience_years }} лет</p>
        <p v-if="data.qualifications" class="feature-text">Квалификация: {{ data.qualifications }}</p>
        <p v-if="data.rating != null">Рейтинг {{ data.rating }} · Отзывов: {{ data.review_count }}</p>
        <div v-if="data.published_programs?.length" class="feature-stack">
          <h2>Программы тренера</h2>
          <RouterLink v-for="(id, index) in data.published_programs" :key="id" class="feature-link" :to="`/marketplace/programs/${id}`">
            Программа {{ index + 1 }}
          </RouterLink>
        </div>
      </section>

      <form v-if="!isPublic && draft && auth.token" class="feature-card" @submit.prevent="save">
        <span v-if="data?.verified" class="feature-badge">Тренер подтверждён</span>
        <fieldset class="feature-stack" :disabled="saving">
          <legend>Публичные сведения</legend>
          <label>Имя тренера<input v-model.trim="draft.display_name" maxlength="200" required /></label>
          <label>Адрес фотографии<input v-model.trim="draft.photo_ref" maxlength="2048" placeholder="https://…" /></label>
          <label>О тренере<textarea v-model.trim="draft.bio" maxlength="10000" /></label>
          <label>Специализации, по одной на строку<textarea v-model="draft.specializations" /></label>
          <label>Опыт, лет<input v-model.number="draft.experience_years" type="number" min="0" required /></label>
          <label>Квалификация<textarea v-model.trim="draft.qualifications" maxlength="10000" /></label>
        </fieldset>
        <button class="feature-button" :disabled="saving">{{ saving ? 'Сохраняем…' : 'Сохранить профиль тренера' }}</button>
        <p v-if="saved" class="feature-success" role="status">Профиль тренера сохранён</p>
        <RequestState :error="saveError" :status="saveStatus" />
        <div v-if="data?.id" class="feature-row">
          <RouterLink class="feature-link" :to="`/trainers/${data.id}`">Публичный профиль</RouterLink>
          <RouterLink class="feature-link" to="/trainer/verification">Проверка квалификации</RouterLink>
          <RouterLink class="feature-link" to="/trainer/programs">Мои программы</RouterLink>
        </div>
      </form>
    </div>
  </FeaturePage>
</template>
