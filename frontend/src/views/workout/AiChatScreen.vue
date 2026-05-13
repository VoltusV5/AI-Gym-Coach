<template>
  <ion-page class="ai-chat-page">
    <ion-header class="ion-no-border chat-header">
      <ion-toolbar class="chat-toolbar">
        <ion-buttons slot="start">
          <ion-back-button default-href="/workout-tools" text="" color="dark"></ion-back-button>
        </ion-buttons>
        <div class="header-center">
          <ion-title class="chat-title">AI Тренер</ion-title>
          <span class="chat-status">онлайн</span>
        </div>
      </ion-toolbar>
    </ion-header>

    <ion-content class="chat-content" fullscreen>
      <div class="chat-messages-container">
        <div v-if="error" class="chat-error-msg">
          {{ error }}
        </div>

        <div v-if="messages.length === 0 && !isLoading" class="chat-welcome">
          <div class="welcome-badge">🤖</div>
          <p class="welcome-title">Привет! Я ваш персональный AI-тренер.</p>
          <p class="welcome-subtitle">Задайте мне любой вопрос по технике выполнения упражнений, подбору весов, замене элементов или составлению программы.</p>
        </div>

        <div class="message-list">
          <div
            v-for="(msg, index) in messages"
            :key="index"
            class="message-wrapper"
            :class="msg.role === 'user' ? 'message-wrapper--user' : 'message-wrapper--assistant'"
          >
            <div class="message-bubble">
              <div
                class="message-text md-rendered-content"
                v-html="renderMarkdown(msg.displayedContent !== undefined ? msg.displayedContent : msg.content)"
              ></div>
              <span v-if="animatingMessageIndex === index" class="blinking-cursor"></span>
            </div>
          </div>

          <div v-if="isLoading" class="message-wrapper message-wrapper--assistant">
            <div class="message-bubble message-bubble--loading">
              <ion-spinner name="dots" class="typing-dots"></ion-spinner>
            </div>
          </div>
        </div>
      </div>
    </ion-content>

    <ion-footer class="ion-no-border chat-footer">
      <form class="chat-input-form" @submit.prevent="sendMessage">
        <div class="input-bar-glass">
          <input
            v-model="inputMessage"
            type="text"
            class="chat-text-input"
            placeholder="Спросите AI тренера..."
            :disabled="isLoading"
          />
          <button
            type="submit"
            class="chat-send-btn"
            :disabled="!inputMessage.trim() || isLoading"
            title="Отправить"
          >
            <ion-icon :icon="sendOutline" class="send-icon" />
          </button>
        </div>
      </form>
    </ion-footer>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'AiChatScreen' })

import { ref, onMounted, nextTick } from 'vue'
import { IonPage, IonHeader, IonToolbar, IonTitle, IonButtons, IonBackButton, IonContent, IonFooter, IonIcon, IonSpinner } from '@ionic/vue'
import { sendOutline } from 'ionicons/icons'
import api from '@/api/api'

const messages = ref([])
const inputMessage = ref('')
const isLoading = ref(false)
const error = ref(null)

let shouldAnimateLastResponse = false
const animatingMessageIndex = ref(-1)
let animationInterval = null

onMounted(() => {
  fetchHistory()
})

async function fetchHistory() {
  try {
    const { data } = await api.get('/api/v1/ai_chat/history')
    const rawMessages = Array.isArray(data) ? data : []
    messages.value = rawMessages.map(m => ({
      ...m,
      displayedContent: m.content
    }))

    if (shouldAnimateLastResponse && messages.value.length > 0) {
      shouldAnimateLastResponse = false
      const lastIdx = messages.value.length - 1
      if (messages.value[lastIdx].role === 'assistant') {
        startTypingAnimation(lastIdx)
      }
    }
    scrollToBottom()
  } catch (err) {
    console.error('Failed to fetch chat history', err)
    error.value = 'Не удалось загрузить историю сообщений'
  }
}

function startTypingAnimation(index) {
  if (animationInterval) clearInterval(animationInterval)

  animatingMessageIndex.value = index
  const fullText = messages.value[index].content || ''
  messages.value[index].displayedContent = ''
  let currentCharIdx = 0

  const charsPerTick = Math.max(1, Math.floor(fullText.length / 80))

  animationInterval = setInterval(() => {
    currentCharIdx += charsPerTick
    if (currentCharIdx >= fullText.length) {
      currentCharIdx = fullText.length
      clearInterval(animationInterval)
      animationInterval = null
      animatingMessageIndex.value = -1
    }
    messages.value[index].displayedContent = fullText.substring(0, currentCharIdx)
    scrollToBottom()
  }, 15)
}

async function sendMessage() {
  const txt = inputMessage.value.trim()
  if (!txt || isLoading.value) return
  messages.value.push({ role: 'user', content: txt, displayedContent: txt })
  inputMessage.value = ''
  scrollToBottom()

  isLoading.value = true
  error.value = null

  try {
    await api.post('/api/v1/ai_chat/generate_answer', {
      model: 'llama3.1-8b',
      messages: messages.value.map(m => ({ role: m.role, content: m.content }))
    })

    shouldAnimateLastResponse = true
    await fetchHistory()
  } catch (err) {
    console.error('AI chat processing failed', err)
    error.value = 'Ошибка при получении ответа от AI тренера'
  } finally {
    isLoading.value = false
    scrollToBottom()
  }
}

function renderMarkdown(text) {
  if (!text) return ''

  let html = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/\*(.*?)\*/g, '<em>$1</em>')
  html = html.replace(/^### (.*$)/gim, '<h3 class="md-header">$1</h3>')
  html = html.replace(/^## (.*$)/gim, '<h2 class="md-header">$1</h2>')
  html = html.replace(/^[•\-] (.*$)/gim, '<li class="md-list-item">$1</li>')
  html = html.replace(/\n/g, '<br>')

  return html
}

function scrollToBottom() {
  nextTick(() => {
    setTimeout(() => {
      const container = document.querySelector('.chat-messages-container')
      if (container) {
        container.scrollTop = container.scrollHeight
      }
    }, 50)
  })
}

</script>

<style scoped>
.ai-chat-page {
  --background: var(--sportik-bg);
  background: var(--sportik-bg);
}

.chat-header {
  box-shadow: none;
  background: var(--sportik-surface);
  border-bottom: 1px solid var(--sportik-border);
}

.chat-toolbar {
  --background: transparent;
  --border-width: 0;
  --min-height: 56px;
  display: flex;
  align-items: center;
}

.header-center {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
}

.chat-title {
  font-weight: 700;
  font-size: 1.05rem;
  color: var(--sportik-text);
  padding: 0;
  line-height: 1.2;
}

.chat-status {
  font-size: 0.72rem;
  color: var(--sportik-brand);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  line-height: 1;
  margin-top: 2px;
}

.chat-content {
  --background: var(--sportik-bg);
  height: 100%;
}

.chat-messages-container {
  height: 100%;
  overflow-y: auto;
  padding: 16px 16px 24px;
  display: flex;
  flex-direction: column;
}

.chat-error-msg {
  background: color-mix(in srgb, var(--sportik-error, #ef4444) 15%, transparent);
  border: 1px solid color-mix(in srgb, var(--sportik-error, #ef4444) 40%, transparent);
  color: var(--sportik-error, #ef4444);
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 0.88rem;
  text-align: center;
  margin-bottom: 16px;
}

.chat-welcome {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin: auto 0;
  padding: 24px 16px;
}

.welcome-badge {
  font-size: 3rem;
  margin-bottom: 12px;
  animation: floatBounce 3s ease-in-out infinite;
}

@keyframes floatBounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-6px); }
}

.welcome-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--sportik-text);
  margin: 0 0 8px;
}

.welcome-subtitle {
  font-size: 0.88rem;
  color: var(--sportik-text-soft);
  line-height: 1.45;
  margin: 0;
  max-width: 280px;
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: auto;
}

.message-wrapper {
  display: flex;
  width: 100%;
}

.message-wrapper--user {
  justify-content: flex-end;
}

.message-wrapper--assistant {
  justify-content: flex-start;
}

.message-bubble {
  max-width: 82%;
  padding: 12px 16px;
  border-radius: 20px;
  box-shadow: var(--sportik-shadow-sm);
  word-break: break-word;
  transition: transform 0.2s ease;
}

.message-wrapper--user .message-bubble {
  background: linear-gradient(135deg, var(--sportik-brand) 0%, var(--sportik-brand-2) 100%);
  color: #ffffff;
  border-bottom-right-radius: 4px;
}

.message-wrapper--assistant .message-bubble {
  background: var(--sportik-surface);
  color: var(--sportik-text);
  border: 1px solid var(--sportik-border);
  border-bottom-left-radius: 4px;
}

.message-text {
  margin: 0;
  font-size: 0.95rem;
  line-height: 1.4;
  white-space: pre-wrap;
}

.message-bubble--loading {
  padding: 10px 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.typing-dots {
  --color: var(--sportik-brand);
  height: 24px;
  width: 36px;
}

.chat-footer {
  background: transparent;
  padding: 8px 16px calc(12px + env(safe-area-inset-bottom, 0px));
}

.input-bar-glass {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--sportik-surface-glass);
  backdrop-filter: blur(16px);
  border: 1px solid var(--sportik-border);
  border-radius: 24px;
  padding: 6px 6px 6px 16px;
  box-shadow: var(--sportik-shadow-md);
}

.chat-text-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--sportik-text);
  font-size: 0.95rem;
  padding: 8px 0;
  font-family: inherit;
}

.chat-text-input::placeholder {
  color: var(--sportik-text-muted);
}

.chat-send-btn {
  background: linear-gradient(135deg, var(--sportik-brand) 0%, var(--sportik-brand-2) 100%);
  border: none;
  border-radius: 50%;
  width: 38px;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  cursor: pointer;
  transition: transform 0.15s ease, opacity 0.15s ease;
  flex-shrink: 0;
}

.chat-send-btn:active {
  transform: scale(0.92);
}

.chat-send-btn:disabled {
  opacity: 0.4;
  cursor: default;
  transform: none;
}

.send-icon {
  font-size: 1.15rem;
  margin-left: 1px;
}


.message-bubble :deep(strong) {
  font-weight: 700;
  color: inherit;
}

.message-bubble :deep(em) {
  font-style: italic;
}

.message-bubble :deep(.md-header) {
  font-size: 1.05rem;
  font-weight: 700;
  margin: 10px 0 6px;
  line-height: 1.25;
  color: inherit;
}

.message-bubble :deep(.md-list-item) {
  margin-left: 16px;
  list-style-type: disc;
  margin-bottom: 4px;
  line-height: 1.4;
}

.blinking-cursor {
  display: inline-block;
  width: 6px;
  height: 15px;
  background-color: var(--sportik-brand);
  vertical-align: middle;
  margin-left: 4px;
  border-radius: 2px;
  animation: cursorBlink 0.8s infinite;
}

@keyframes cursorBlink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}
</style>