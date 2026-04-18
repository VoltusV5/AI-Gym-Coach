<template>
  <ion-page class="notes-page">
    <ion-content class="notes-content" fullscreen>
      <div class="notes-scroll">
        <div class="notes-frame">
          <div class="notes-apollo-strip" aria-hidden="true">
            <img v-if="apolloSrc" class="notes-apollo-img" :src="apolloSrc" alt="" />
          </div>

          <div class="notes-sheet">
            <div class="notes-sheet-inner ion-padding">
              <h1 class="notes-title">{{ t.title }}</h1>

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
                  <p v-if="preview(note.body)" class="note-card-preview">{{ preview(note.body) }}</p>
                  <p class="note-card-date">{{ formatDate(note.createdAt) }}</p>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </ion-content>

    <ion-footer class="ion-no-border notes-page-footer">
      <div class="notes-footer-stack">
        <app-tab-bar active-key="notes" />
      </div>
    </ion-footer>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'NotesListPage' })

import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { IonPage, IonContent, IonFooter, IonButton } from '@ionic/vue'
import { useNotesStore } from '@/stores/notes'
import AppTabBar from '@/components/navigation/AppTabBar.vue'
import { getApolloHeaderImageUrl } from '@/utils/localImages'

const apolloSrc = getApolloHeaderImageUrl()

const t = {
  title: '\u0417\u0430\u043c\u0435\u0442\u043a\u0438',
  create: '\u0421\u043e\u0437\u0434\u0430\u0442\u044c \u0437\u0430\u043c\u0435\u0442\u043a\u0443',
  empty: '\u041f\u043e\u043a\u0430 \u0437\u0430\u043c\u0435\u0442\u043e\u043a \u043d\u0435\u0442. \u0421\u043e\u0437\u0434\u0430\u0439\u0442\u0435 \u043d\u043e\u0432\u0443\u044e \u0437\u0430\u043c\u0435\u0442\u043a\u0443',
  noTitle: '\u0411\u0435\u0437 \u043d\u0430\u0437\u0432\u0430\u043d\u0438\u044f',
  unknownDate: '\u0414\u0430\u0442\u0430 \u043d\u0435\u0438\u0437\u0432\u0435\u0441\u0442\u043d\u0430'
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
  if (!txt) return ''
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
.notes-content { --background: var(--sportik-bg); }
.notes-scroll {
  padding-bottom: 0;
  background: transparent;
}
.notes-frame {
  --notes-apollo-h: clamp(124px, 31vw, 176px);
  --notes-footer-pad: calc(120px + env(safe-area-inset-bottom, 0px));
  min-height: calc(100svh - env(safe-area-inset-bottom, 0px));
  width: 100%;
  display: flex;
  flex-direction: column;
}
.notes-apollo-strip {
  flex: 0 0 var(--notes-apollo-h);
  width: 100%;
  overflow: hidden;
  position: relative;
  background: linear-gradient(145deg, var(--sportik-brand) 0%, var(--sportik-brand-2) 100%);
}
.notes-apollo-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 18%;
  display: block;
}
.notes-sheet {
  flex: 1 1 auto;
  width: 100%;
  min-width: 0;
  margin-top: -16px;
  position: relative;
  z-index: 1;
  border-radius: var(--sportik-radius-xl) var(--sportik-radius-xl) 0 0;
  background: var(--sportik-surface);
  min-height: calc(100svh - var(--notes-apollo-h) - env(safe-area-inset-bottom, 0px) + 8px);
  padding-bottom: calc(var(--notes-footer-pad) + 4px);
  color: var(--sportik-text);
  box-shadow: var(--sportik-shadow-lg);
  transform: translateZ(0);
}
.notes-sheet-inner {
  padding-top: 1rem;
  padding-bottom: 0.25rem;
}
.notes-title { margin: 0 0 1rem; text-align: center; font-size: clamp(1.5rem, 5vw, 2rem); font-weight: 700; color: var(--sportik-text); }
.notes-create { margin-bottom: 0.9rem; }
.notes-empty { margin: 0.35rem 0 0.9rem; text-align: center; color: var(--sportik-text-muted); font-size: 0.9rem; }
.notes-list { display: flex; flex-direction: column; gap: 10px; }
.note-card { border: 1px solid var(--sportik-border); width: 100%; text-align: left; background: var(--sportik-surface-soft); border-radius: 14px; padding: 12px; cursor: pointer; color: var(--sportik-text); box-shadow: var(--sportik-shadow-md); }
.note-card-title { margin: 0; font-size: 1rem; font-weight: 700; color: var(--sportik-text); }
.note-card-preview { margin: 6px 0 8px; font-size: 0.9rem; line-height: 1.35; color: var(--sportik-text-soft); }
.note-card-date { margin: 0; font-size: 0.78rem; color: var(--sportik-text-muted); }
.notes-page-footer { box-shadow: 0 -8px 22px rgba(0, 0, 0, 0.08); }
.notes-footer-stack { background: var(--sportik-surface-glass); backdrop-filter: blur(12px); padding-bottom: env(safe-area-inset-bottom, 0px); }
</style>
