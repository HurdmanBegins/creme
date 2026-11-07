<script setup>
import { ref } from 'vue'
import { api, setToken } from '../api'
import Mascot from './Mascot.vue'

const emit = defineEmits(['success'])
const password = ref('')
const error = ref('')
const busy = ref(false)
const shake = ref(false)

async function submit() {
  if (!password.value || busy.value) return
  busy.value = true
  error.value = ''
  try {
    const { token } = await api.login(password.value)
    setToken(token)
    emit('success')
  } catch (e) {
    error.value = e.message || 'Oups, petit souci côté cuisine'
    password.value = ''
    shake.value = true
    setTimeout(() => (shake.value = false), 500)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <section class="login card" :class="{ shake }">
    <Mascot :mood="password ? 'shy' : 'happy'" class="mascot-lg" />
    <h1 class="logo">Crême <span class="version">2.0</span></h1>
    <p class="tagline">Psst… le mot de passe ? 🤫</p>
    <form @submit.prevent="submit">
      <input
        v-model="password"
        type="password"
        placeholder="mot de passe secret"
        autocomplete="current-password"
        autofocus
      />
      <button class="btn btn-primary" type="submit" :disabled="busy || !password">
        {{ busy ? 'On vérifie…' : 'Entrer 🍦' }}
      </button>
    </form>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
  </section>
</template>
