<template>
  <ion-page class="notes-page">
    <ion-content class="notes-content" fullscreen>
      <div class="notes-frame">
        <div class="notes-strip" aria-hidden="true"></div>

        <div class="notes-sheet">
          <div class="notes-inner ion-padding">
            <h1 class="notes-title">{{ t.title }}</h1>
            <p class="notes-subtitle">{{ t.subtitle }}</p>

            <ion-button class="sportik-footer-btn notes-create" expand="block" @click="onCreateNote">
              {{ t.create }}
            </ion-button>

            <p v-if="!items.length" class="notes-empty">
              {{ t.empty }}
            </p>

            <div class="notes-list">
              <button
                v-for="note in items"
                :key="note.id"
                type="button"
                class="note-card"
                @click="openNote(note.id)"
              >
                <p class="note-card-title">{{ note.title || t.noTitle }}</p>
                <p class="note-card-preview">{{ preview(note.body) }}</p>
                <p class="note-card-date">{{ formatDate(note.createdAt) }}</p>
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="notes-footer-stack">
        <app-tab-bar active-key="notes" />
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'NotesListPage' })

import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { IonPage, IonContent, IonButton } from '@ionic/vue'
import { useNotesStore } from '@/stores/notes'
import AppTabBar from '@/components/navigation/AppTabBar.vue'

const t = {
  title: '\u0417\u0430\u043c\u0435\u0442\u043a\u0438',
  subtitle: '\u0411\u044b\u0441\u0442\u0440\u044b\u0435 \u0437\u0430\u043f\u0438\u0441\u0438 \u043f\u043e \u0442\u0440\u0435\u043d\u0438\u0440\u043e\u0432\u043a\u0430\u043c \u0438 \u043f\u0440\u043e\u0433\u0440\u0435\u0441\u0441\u0443',
  create: '\u0421\u043e\u0437\u0434\u0430\u0442\u044c \u0437\u0430\u043c\u0435\u0442\u043a\u0443',
  empty: '\u041f\u043e\u043a\u0430 \u0437\u0430\u043c\u0435\u0442\u043e\u043a \u043d\u0435\u0442. \u0421\u043e\u0437\u0434\u0430\u0439\u0442\u0435 \u043f\u0435\u0440\u0432\u0443\u044e \u0437\u0430\u043c\u0435\u0442\u043a\u0443 - \u0437\u0430\u0433\u043e\u043b\u043e\u0432\u043e\u043a, \u0442\u0435\u043a\u0441\u0442 \u0438 \u0434\u0430\u0442\u0430 \u0431\u0443\u0434\u0443\u0442 \u0441\u043e\u0445\u0440\u0430\u043d\u0435\u043d\u044b.',
  noTitle: '\u0411\u0435\u0437 \u043d\u0430\u0437\u0432\u0430\u043d\u0438\u044f',
  unknownDate: '\u0414\u0430\u0442\u0430 \u043d\u0435\u0438\u0437\u0432\u0435\u0441\u0442\u043d\u0430',
  emptyBody: '\u041d\u0430\u0436\u043c\u0438\u0442\u0435, \u0447\u0442\u043e\u0431\u044b \u0434\u043e\u0431\u0430\u0432\u0438\u0442\u044c \u0442\u0435\u043a\u0441\u0442 \u0437\u0430\u043c\u0435\u0442\u043a\u0438.'
}

const router = useRouter()
const notesStore = useNotesStore()

onMounted(() => {
  notesStore.hydrate()
  notesStore.syncWithApi()
})

const items = computed(() => notesStore.sortedNotes)

function formatDate(value) {
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return t.unknownDate
  return d.toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function preview(body) {
  const txt = String(body ?? '').trim()
  if (!txt) return t.emptyBody
  return txt.length > 110 ? `${txt.slice(0, 110)}...` : txt
}

function openNote(id) {
  router.push(`/notes/${id}`)
}

function onCreateNote() {
  const note = notesStore.createEmptyDraft()
  router.push(`/notes/${note.id}`)
}
</script>

<style scoped>
.notes-content { --background: var(--sportik-mint-soft); }
.notes-frame { min-height: calc(100svh - env(safe-area-inset-bottom, 0px)); display: flex; flex-direction: column; }
.notes-strip { flex: 0 0 clamp(108px, 26vw, 152px); background: linear-gradient(165deg, #b8fcff 0%, var(--sportik-cyan) 50%, #52e8e8 100%); }
.notes-sheet { flex: 1; margin-top: -8px; border-radius: 28px 28px 0 0; background: var(--sportik-cream); padding-bottom: calc(120px + env(safe-area-inset-bottom, 0px)); color: var(--sportik-text); }
.notes-title { margin: 0; text-align: center; font-family: 'Roboto Mono', 'Roboto', monospace; font-size: clamp(1.65rem, 5vw, 2.2rem); color: var(--sportik-text); }
.notes-subtitle { text-align: center; margin: 0.4rem 0 1rem; font-size: 0.9rem; color: var(--sportik-text-muted); }
.notes-create { margin-bottom: 0.9rem; }
.notes-empty { margin: 0.35rem 0 0.9rem; text-align: center; color: var(--sportik-text-muted); font-size: 0.9rem; }
.notes-list { display: flex; flex-direction: column; gap: 10px; }
.note-card { border: none; width: 100%; text-align: left; background: var(--sportik-card-gray); border-radius: 12px; padding: 12px; cursor: pointer; color: var(--sportik-text); }
.note-card-title { margin: 0; font-size: 1rem; font-weight: 700; color: var(--sportik-text); }
.note-card-preview { margin: 6px 0 8px; font-size: 0.9rem; line-height: 1.35; color: var(--sportik-text-soft); }
.note-card-date { margin: 0; font-size: 0.78rem; color: var(--sportik-text-muted); }
.notes-footer-stack { position: fixed; left: 0; right: 0; bottom: 0; z-index: 10; background: var(--sportik-cream); box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.06); padding-bottom: env(safe-area-inset-bottom, 0px); }
</style>
