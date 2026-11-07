<script setup>
import { ref, onMounted } from 'vue'
import { api, getToken, setToken } from './api'
import LoginView from './components/LoginView.vue'
import Dashboard from './components/Dashboard.vue'
import Mascot from './components/Mascot.vue'
import { applyTheme } from './composables/useTheme'

applyTheme()
const state = ref('checking') // checking | anonymous | authed

const COLORS = ['#ff8fb8', '#c78bff', '#7fd8b8', '#7fc4ff', '#ffb86b', '#ffe066']
const sprinkles = Array.from({ length: 22 }, (_, i) => ({
  left: `${(i * 37 + 11) % 100}%`,
  top: `${(i * 53 + 7) % 100}%`,
  background: COLORS[i % COLORS.length],
  transform: `rotate(${(i * 47) % 180}deg)`,
  animationDelay: `${(i % 7) * -1.3}s`,
  animationDuration: `${8 + (i % 5) * 2}s`,
}))

onMounted(async () => {
  if (!getToken()) {
    state.value = 'anonymous'
    return
  }
  try {
    await api.me()
    state.value = 'authed'
  } catch {
    setToken(null)
    state.value = 'anonymous'
  }
})

function logout() {
  setToken(null)
  state.value = 'anonymous'
}
</script>

<template>
  <div class="sprinkles" aria-hidden="true">
    <span v-for="(s, i) in sprinkles" :key="i" class="sprinkle" :style="s"></span>
  </div>

  <main class="page">
    <div v-if="state === 'checking'" class="center loading">
      <Mascot mood="happy" class="mascot-lg" />
      <p>On fouette la crème… 🥣</p>
    </div>
    <LoginView v-else-if="state === 'anonymous'" @success="state = 'authed'" />
    <Dashboard v-else @logout="logout" @expired="logout" />
  </main>
</template>
