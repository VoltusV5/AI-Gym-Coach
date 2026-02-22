<template>
  <ion-app>
    <router-view />
  </ion-app>
</template>

<style>
html, body, #app, ion-app {
  height: 100%;
  margin: 0;
  padding: 0;
  overflow: hidden;
}

ion-app {
  display: flex;
  flex-direction: column;
}
</style>