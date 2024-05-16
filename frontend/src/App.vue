<script lang="ts" setup>
import { onMounted } from 'vue'
// import { InitDatabase } from '../wailsjs/go/main/App'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

onMounted(async () => {
  if (!localStorage.getItem('locale')) {
    localStorage.setItem('locale', 'en-US')
  }
  locale.value = localStorage.getItem('locale') as string

  if (!localStorage.getItem('theme')) {
    localStorage.setItem('theme', 'light')
  }
  document.documentElement.className = localStorage.getItem('theme') as string
  
  // await InitDatabase()
})
</script>

<template>
  <router-view />
</template>

<style lang="scss">
#app {
  display: flex;
}

p {
  margin: 0;
}
</style>