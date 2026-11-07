<script setup>
import { computed, ref, watch } from 'vue'
import { now } from '../composables/useClock'
import {
  elapsedMs, splitDuration, moodFor, formatDateTime, formatShort, pad,
  hasGain, gainValue, gainDecimals, formatNumber, gainEmoji, PERIOD_LABELS,
} from '../utils/time'
import { durationMilestone, gainMilestone, getSeen, setSeen } from '../utils/milestones'
import { equivalences, pickEquivalence } from '../utils/equivalences'
import { renderCardPng, shareOrDownload } from '../utils/share'

const props = defineProps({
  counter: { type: Object, required: true },
})
const emit = defineEmits(['edit', 'delete', 'toggle-pause', 'reset', 'archive', 'unarchive', 'milestone', 'notify'])

const confirming = ref(null) // 'delete' | 'reset' | null
const sharing = ref(false)

const elapsed = computed(() => elapsedMs(props.counter, now.value))
const future = computed(() => elapsed.value < 0)
const parts = computed(() => splitDuration(Math.abs(elapsed.value)))
const mood = computed(() => moodFor(elapsed.value, props.counter.paused))

// ----- gain + équivalence -----
const gain = computed(() => hasGain(props.counter))
const gainNumber = computed(() => gainValue(props.counter, elapsed.value))
const gainText = computed(() => formatNumber(gainNumber.value, gainDecimals(props.counter)))
const gainIcon = computed(() => gainEmoji(props.counter.gainUnit))
const rateText = computed(() => {
  const amount = Number(props.counter.gainAmount)
  const period = PERIOD_LABELS[props.counter.gainPeriod] || 'jour'
  return `${formatNumber(amount, Number.isInteger(amount) ? 0 : 2)} ${props.counter.gainUnit} / ${period}`
})
const equivText = computed(() =>
  gain.value ? pickEquivalence(equivalences(props.counter, gainNumber.value), now.value) : '',
)

// ----- tentatives (« j'ai craqué ») -----
const attempts = computed(() => props.counter.attempts || [])
const bestPreviousMs = computed(() => attempts.value.reduce((m, a) => Math.max(m, a.seconds * 1000), 0))
const isRecord = computed(() => attempts.value.length > 0 && elapsed.value > bestPreviousMs.value)
const recordMs = computed(() => Math.max(bestPreviousMs.value, elapsed.value))

// ----- paliers -----
const dm = computed(() => durationMilestone(Math.max(0, elapsed.value)))
const gm = computed(() => (gain.value ? gainMilestone(gainNumber.value) : { reachedIndex: -1, reached: null }))
const milestoneKey = computed(() => `${dm.value.reachedIndex}|${gm.value.reachedIndex}`)

function parseKey(key) {
  const [d, g] = String(key ?? '-1|-1').split('|').map(Number)
  return { d, g }
}

// On ne fête qu'un palier franchi depuis la dernière fois qu'on a vu ce compteur.
watch(
  milestoneKey,
  (key) => {
    const seen = getSeen(props.counter.id)
    setSeen(props.counter.id, key)
    if (seen === null || props.counter.archived || props.counter.paused || future.value) return
    const before = parseKey(seen)
    const after = parseKey(key)
    if (after.d > before.d && dm.value.reached) {
      emit('milestone', { counter: props.counter, phrase: dm.value.reached.phrase })
    } else if (after.g > before.g && gm.value.reached) {
      const text = `${formatNumber(gm.value.reached, 0)} ${props.counter.gainUnit} ${props.counter.gainLabel}`.replace(/\s+/g, ' ').trim()
      emit('milestone', { counter: props.counter, phrase: `${text} ! Ça commence à chiffrer 🤑` })
    }
  },
  { immediate: true },
)

// ----- carte à partager -----
async function share() {
  if (sharing.value) return
  sharing.value = true
  try {
    const blob = await renderCardPng(props.counter, {
      elapsedMs: elapsed.value,
      gainText: gain.value ? `${gainText.value} ${props.counter.gainUnit} ${props.counter.gainLabel}`.replace(/\s+/g, ' ').trim() : '',
      equivText: equivText.value,
      mood: mood.value,
    })
    const slug = props.counter.title.toLowerCase().normalize('NFD').replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') || 'compteur'
    const result = await shareOrDownload(blob, `creme-${slug}.png`, props.counter.title)
    if (result === 'downloaded') emit('notify', 'Carte téléchargée, à toi de frimer 📤')
    if (result === 'shared') emit('notify', 'Carte partagée 💌')
  } catch {
    emit('notify', 'Impossible de fabriquer la carte 🫠')
  } finally {
    sharing.value = false
  }
}
</script>

<template>
  <article class="card counter" :class="[`c-${counter.color}`, { paused: counter.paused, archived: counter.archived }]">
    <header class="counter-head">
      <span class="avatar" aria-hidden="true">{{ counter.emoji }}</span>
      <div class="titles">
        <h2>{{ counter.title }}</h2>
        <p v-if="counter.phrase" class="phrase">{{ counter.phrase }}</p>
      </div>
      <span v-if="counter.archived" class="badge">📦 archivé</span>
      <span v-else-if="counter.paused" class="badge">💤 pause</span>
    </header>

    <p v-if="attempts.length" class="attempts">
      🔁 Tentative n°{{ attempts.length + 1 }} · record {{ formatShort(recordMs) }}
      <span v-if="isRecord" class="record">🏆 en cours !</span>
    </p>

    <p class="since-label">{{ future ? 'Ça commence dans' : 'Depuis' }}</p>
    <div class="clock" :class="{ wide: parts.years > 0 }">
      <span v-if="parts.years" class="unit">
        <b>{{ parts.years }}</b><small>{{ parts.years > 1 ? 'ans' : 'an' }}</small>
      </span>
      <span class="unit"><b>{{ parts.days }}</b><small>{{ parts.days > 1 ? 'jours' : 'jour' }}</small></span>
      <span class="unit"><b>{{ pad(parts.hours) }}</b><small>h</small></span>
      <span class="unit"><b>{{ pad(parts.minutes) }}</b><small>min</small></span>
      <span class="unit unit-sec">
        <b :key="parts.seconds" class="pop">{{ pad(parts.seconds) }}</b><small>s</small>
      </span>
    </div>

    <div v-if="gain" class="gain">
      <span v-if="gainIcon" class="gain-icon" aria-hidden="true">{{ gainIcon }}</span>
      <span class="gain-text"><b>{{ gainText }}</b> {{ counter.gainUnit }} {{ counter.gainLabel }}</span>
      <small class="gain-rate">{{ rateText }}</small>
      <small v-if="attempts.length" class="gain-note">cumul de {{ attempts.length + 1 }} tentatives, rien n'est perdu 💪</small>
      <Transition name="fade" mode="out-in">
        <span v-if="equivText" :key="equivText" class="equiv">{{ equivText }}</span>
      </Transition>
    </div>

    <div v-if="!future && !counter.archived" class="milestone">
      <div class="milestone-row">
        <span class="milestone-reached">{{ dm.reached ? '🏅 ' + dm.reached.label : '🥚 premier palier bientôt' }}</span>
        <span v-if="dm.next" class="milestone-next">{{ dm.next.label }} dans {{ formatShort(dm.remainingMs) }}</span>
      </div>
      <div class="progress" role="progressbar" :aria-valuenow="Math.round(dm.progress * 100)" aria-valuemin="0" aria-valuemax="100">
        <div class="progress-bar" :style="{ width: dm.progress * 100 + '%' }"></div>
      </div>
    </div>

    <p class="mood">{{ mood }}</p>

    <footer class="counter-foot">
      <time :datetime="counter.startAt">📅 {{ formatDateTime(counter.startAt) }}</time>

      <div v-if="counter.archived" class="actions">
        <template v-if="confirming === 'delete'">
          <span class="confirm">Supprimer pour de bon ? 🫠</span>
          <button class="btn btn-danger btn-sm" @click="emit('delete')">Oui</button>
          <button class="btn btn-sm" @click="confirming = null">Non</button>
        </template>
        <template v-else>
          <button class="btn btn-sm" @click="emit('unarchive')">↩️ Restaurer</button>
          <button class="btn btn-sm" title="Supprimer définitivement" @click="confirming = 'delete'">🗑️ Supprimer</button>
        </template>
      </div>

      <div v-else-if="confirming === 'reset'" class="actions">
        <span class="confirm">Tout repart de zéro ? On garde le souvenir 🫠</span>
        <button class="btn btn-danger btn-sm" @click="confirming = null; emit('reset')">Oui, j'ai craqué</button>
        <button class="btn btn-sm" @click="confirming = null">Non !</button>
      </div>

      <div v-else class="actions">
        <button class="btn btn-sm" :title="counter.paused ? 'Reprendre' : 'Mettre en pause'" @click="emit('toggle-pause')">
          {{ counter.paused ? '▶️ Reprendre' : '⏸️ Pause' }}
        </button>
        <button v-if="!future" class="btn btn-sm" title="Repartir de zéro en gardant l'historique" @click="confirming = 'reset'">🫠 J'ai craqué</button>
        <button class="btn btn-sm" title="Modifier" @click="emit('edit')">✏️</button>
        <button class="btn btn-sm" title="Partager en image" :disabled="sharing" @click="share">📤</button>
        <button class="btn btn-sm" title="Archiver" @click="emit('archive')">📦</button>
      </div>
    </footer>
  </article>
</template>
