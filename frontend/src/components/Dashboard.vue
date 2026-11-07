<script setup>
import { ref, computed, onMounted } from 'vue'
import { api, ApiError } from '../api'
import { setServerTime } from '../composables/useClock'
import { useTheme } from '../composables/useTheme'
import { forgetSeen } from '../utils/milestones'
import CounterCard from './CounterCard.vue'
import CounterForm from './CounterForm.vue'
import Confetti from './Confetti.vue'
import Mascot from './Mascot.vue'

const emit = defineEmits(['logout', 'expired'])
const { isDark, toggleTheme } = useTheme()

const TAGLINES = [
  'Le compteur qui monte comme de la chantilly',
  'Depuis quand ? On compte pour vous, seconde par seconde',
  '100 % matière grasse, 0 % stress',
  'Parce que le temps, ça fond vite',
  'Mesure tout, surtout ce qui ne sert à rien',
]
const tagline = TAGLINES[Math.floor(Math.random() * TAGLINES.length)]

const counters = ref([])
const loading = ref(true)
const toast = ref('')
const form = ref({ open: false, counter: null })
const showArchived = ref(false)
const confetti = ref(null)

const active = computed(() => counters.value.filter((c) => !c.archived))
const archived = computed(() => counters.value.filter((c) => c.archived))
const shown = computed(() => (showArchived.value ? archived.value : active.value))
const mascotMood = computed(() => (active.value.length && active.value.every((c) => c.paused) ? 'sleepy' : 'happy'))

// ----- toasts (file d'attente pour les paliers) -----
let toastTimer
const queue = []
function notify(msg, ms = 3500) {
  if (!msg) return
  if (toast.value) {
    queue.push([msg, ms])
    return
  }
  toast.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toast.value = ''
    const next = queue.shift()
    if (next) setTimeout(() => notify(...next), 250)
  }, ms)
}

function onMilestone({ phrase }) {
  confetti.value?.burst()
  notify(`🎉 ${phrase}`, 6000)
}

function handleError(e) {
  if (e instanceof ApiError && e.status === 401) {
    emit('expired')
    return
  }
  notify(e.message || 'Oups, la crème a débordé 🫠')
}

async function load() {
  try {
    const data = await api.list()
    counters.value = data.counters
    setServerTime(data.serverTime)
  } catch (e) {
    handleError(e)
  } finally {
    loading.value = false
  }
}
onMounted(load)

function replace(updated) {
  const i = counters.value.findIndex((c) => c.id === updated.id)
  if (i >= 0) counters.value[i] = updated
}

function openCreate() {
  form.value = { open: true, counter: null }
}
function openEdit(counter) {
  form.value = { open: true, counter }
}
function closeForm() {
  form.value = { open: false, counter: null }
}

async function save(payload) {
  try {
    if (form.value.counter) {
      replace(await api.update(form.value.counter.id, payload))
      notify('Compteur retouché ✨')
    } else {
      counters.value.push(await api.create(payload))
      notify('Nouveau compteur, tout frais 🍦')
    }
    closeForm()
  } catch (e) {
    handleError(e)
  }
}

async function togglePause(counter) {
  try {
    if (counter.paused) {
      replace(await api.resume(counter.id))
      notify('Et ça repart ! 🏃')
    } else {
      replace(await api.pause(counter.id))
      notify('Chut, la crème fait la sieste 💤')
    }
  } catch (e) {
    handleError(e)
  }
}

async function reset(counter) {
  try {
    replace(await api.reset(counter.id))
    forgetSeen(counter.id)
    notify("Ça arrive aux meilleurs. On repart, le souvenir est gardé 🔁")
  } catch (e) {
    handleError(e)
  }
}

async function archive(counter) {
  try {
    replace(await api.archive(counter.id))
    notify('Rangé dans la boîte, récupérable dans les archives 📦')
  } catch (e) {
    handleError(e)
  }
}

async function unarchive(counter) {
  try {
    replace(await api.unarchive(counter.id))
    notify('De retour sur le plateau ↩️')
    if (!archived.value.length) showArchived.value = false
  } catch (e) {
    handleError(e)
  }
}

async function remove(counter) {
  try {
    await api.remove(counter.id)
    counters.value = counters.value.filter((c) => c.id !== counter.id)
    forgetSeen(counter.id)
    notify('Pouf, fondu pour de bon 🫠')
    if (!archived.value.length) showArchived.value = false
  } catch (e) {
    handleError(e)
  }
}
</script>

<template>
  <Confetti ref="confetti" />

  <header class="topbar">
    <div class="brand">
      <Mascot :mood="mascotMood" :night="isDark" class="mascot-md" />
      <div>
        <h1 class="logo">Crême <span class="version">2.0</span></h1>
        <p class="tagline">{{ tagline }}</p>
      </div>
    </div>
    <div class="topbar-actions">
      <button class="btn btn-primary" @click="openCreate">＋ Nouveau compteur</button>
      <button class="btn btn-icon" :title="isDark ? 'Passer en mode jour' : 'Passer en mode nuit'" @click="toggleTheme">
        {{ isDark ? '☀️' : '🌙' }}
      </button>
      <button
        class="btn btn-icon"
        :class="{ active: showArchived }"
        :title="showArchived ? 'Retour aux compteurs' : 'Voir les archives'"
        @click="showArchived = !showArchived"
      >
        📦 <span v-if="archived.length" class="count">{{ archived.length }}</span>
      </button>
      <button class="btn btn-ghost" @click="emit('logout')">👋</button>
    </div>
  </header>

  <p v-if="showArchived" class="section-title">📦 Archives — ces compteurs tournent encore, mais au calme.</p>

  <section v-if="loading" class="center loading">
    <p>On sort les compteurs du frigo… 🧊</p>
  </section>

  <section v-else-if="shown.length === 0 && showArchived" class="card empty center">
    <p class="empty-emoji">🍽️</p>
    <h2>Rien dans les archives</h2>
    <p>Tout est encore dans l'assiette.</p>
    <button class="btn" @click="showArchived = false">← Retour aux compteurs</button>
  </section>

  <section v-else-if="shown.length === 0" class="card empty center">
    <p class="empty-emoji">🍮</p>
    <h2>Pas encore de compteur</h2>
    <p>La crème est toute lisse. On en ajoute un ?</p>
    <button class="btn btn-primary" @click="openCreate">＋ Créer mon premier compteur</button>
  </section>

  <section v-else class="grid">
    <CounterCard
      v-for="c in shown"
      :key="c.id"
      :counter="c"
      @edit="openEdit(c)"
      @delete="remove(c)"
      @toggle-pause="togglePause(c)"
      @reset="reset(c)"
      @archive="archive(c)"
      @unarchive="unarchive(c)"
      @milestone="onMilestone"
      @notify="notify"
    />
  </section>

  <footer class="footer">Crême 2.0 — à consommer sans modération 🍦</footer>

  <CounterForm v-if="form.open" :counter="form.counter" @save="save" @close="closeForm" />

  <Transition name="toast">
    <div v-if="toast" class="toast" role="status">{{ toast }}</div>
  </Transition>
</template>
