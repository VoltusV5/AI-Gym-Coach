<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>Главная</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding ion-text-center">
      <h1>Онбординг пройден!</h1>
      <p>Теперь ты можешь пользоваться всеми функциями приложения.</p>
      <ion-button color="danger" @click="resetToken">Выйти (сбросить токен)</ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup>
import { IonPage, IonHeader, IonToolbar, IonTitle, IonContent, IonButton } from '@ionic/vue'
import { removeToken } from '@/utils/auth'

const resetToken = async () => {
  await removeToken()
  window.location.reload()
}
</script>
