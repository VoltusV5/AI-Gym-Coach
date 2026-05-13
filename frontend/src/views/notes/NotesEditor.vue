<template>
  <ion-page class="note-page">
    <ion-content class="note-content" fullscreen>
      <div class="note-frame ion-padding">
        <ion-input
          v-model="title"
          class="note-title-input"
          :placeholder="t.titlePlaceholder"
          maxlength="120"
        />

        <div class="note-body-wrap">
          <ion-textarea
            v-model="body"
            class="note-body-input"
            :placeholder="t.bodyPlaceholder"
            :auto-grow="false"
            rows="12"
          />
        </div>
      </div>

    </ion-content>

    <ion-footer class="ion-no-border note-page-footer">
      <div class="note-bottom-sheet ion-padding">
        <p class="note-date">{{ t.created }}: {{ createdAtLabel }}</p>
        <ion-button class="sportik-footer-btn" expand="block" @click="saveNote">
          {{ t.save }}
        </ion-button>
      </div>
      <div class="note-footer-stack">
        <app-tab-bar active-key="notes" />
      </div>
    </ion-footer>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'NotesEditorPage' })

import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { IonPage, IonContent, IonFooter, IonButton, IonInput, IonTextarea, toastController } from '@ionic/vue'
import AppTabBar from '@/components/navigation/AppTabBar.vue'
import { useNotesStore } from '@/stores/notes'

const t = {
  titlePlaceholder: 'Название заметки...',
  bodyPlaceholder: 'Начните писать здесь...',
  created: 'Создано',
  save: 'Сохранить',
  unknown: 'неизвестно',
  saved: 'Заметка сохранена'
}

const route = useRoute()
const router = useRouter()
const notesStore = useNotesStore()

const title = ref('')
const body = ref('')
const createdAt = ref('')

const noteId = computed(() => String(route.params.id ?? ''))

onMounted(() => {
  notesStore.hydrate()
  const existing = notesStore.noteById(noteId.value)
  if (existing) {
    title.value = existing.title || ''
    body.value = existing.body || ''
    createdAt.value = existing.createdAt
    return
  }

  const draft = notesStore.upsertNote(noteId.value, '', '')
  createdAt.value = draft?.createdAt || new Date().toISOString()
})

const createdAtLabel = computed(() => {
  const d = new Date(createdAt.value)
  if (Number.isNaN(d.getTime())) return t.unknown
  return d.toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
})

async function saveNote() {
  await notesStore.saveNoteWithApi(noteId.value, title.value, body.value)
  await router.push('/notes')
  const toast = await toastController.create({
    message: t.saved,
    duration: 1300,
    color: 'success'
  })
  await toast.present()
}
</script>

<style scoped>
.note-content {
  --background: var(--sportik-bg);
}

.note-frame {
  min-height: min(60vh, calc(100svh - 200px));
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.note-title-input {
  --background: var(--sportik-surface);
  --color: var(--sportik-text);
  --padding-start: 14px;
  --padding-end: 14px;
  border: 1px solid var(--sportik-border);
  box-shadow: var(--sportik-shadow-md);
  border-radius: 14px;
  font-weight: 700;
}

.note-body-wrap {
  flex: 1;
  min-height: 0;
}

.note-body-input {
  --background: var(--sportik-surface);
  --color: var(--sportik-text);
  --padding-top: 12px;
  --padding-bottom: 12px;
  --padding-start: 14px;
  --padding-end: 14px;
  border: 1px solid var(--sportik-border);
  box-shadow: var(--sportik-shadow-md);
  border-radius: 16px;
  height: 100%;
  min-height: 100%;
}

.note-page-footer {
  box-shadow: 0 -8px 22px rgba(0, 0, 0, 0.08);
}

.note-bottom-sheet {
  background: var(--sportik-surface-glass);
  border-top: 1px solid var(--sportik-border);
  backdrop-filter: blur(12px);
}

.note-date {
  margin: 0 0 0.6rem;
  font-size: 0.82rem;
  color: var(--sportik-text-muted);
}

.note-footer-stack {
  background: var(--sportik-surface-glass);
  backdrop-filter: blur(12px);
  padding-bottom: env(safe-area-inset-bottom, 0px);
}
</style>