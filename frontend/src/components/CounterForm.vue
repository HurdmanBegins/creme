<script setup>
import { reactive, ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { toLocalInput, fromLocalInput, PERIOD_LABELS } from '../utils/time'

const props = defineProps({
  counter: { type: Object, default: null }, // null = création
})
const emit = defineEmits(['save', 'close'])

const EMOJIS = ['🍦', '🍰', '🧁', '🍓', '☕', '🚭', '🏃', '💻', '🛏️', '🌸', '🐱', '🎮', '💖', '⭐', '🧘', '🎉']
const COLORS = ['cream', 'pink', 'lavender', 'mint', 'sky', 'peach', 'lemon']
const UNITS = [
  { unit: '€', label: 'économisés' },
  { unit: 'kcal', label: 'évitées' },
  { unit: 'cigarettes', label: 'non fumées' },
  { unit: 'heures', label: 'gagnées' },
  { unit: 'km', label: 'parcourus' },
  { unit: '🍺', label: 'évitées' },
  { unit: '☕', label: 'évités' },
]

const form = reactive({
  emoji: props.counter?.emoji ?? '🍦',
  title: props.counter?.title ?? '',
  phrase: props.counter?.phrase ?? '',
  color: props.counter?.color ?? COLORS[Math.floor(Math.random() * COLORS.length)],
  startAt: toLocalInput(props.counter?.startAt),
  gainAmount: props.counter?.gainAmount > 0 ? String(props.counter.gainAmount).replace('.', ',') : '',
  gainUnit: props.counter?.gainUnit ?? '',
  gainPeriod: props.counter?.gainPeriod || 'day',
  gainLabel: props.counter?.gainLabel ?? '',
  equivUnit: props.counter?.equivUnit ?? '',
  equivValue: props.counter?.equivValue > 0 ? String(props.counter.equivValue).replace('.', ',') : '',
})

// Montant saisi librement : « 13 », « 13,5 », « 13.5 », « 13 € », « 1 200 »… Vide = pas de gain, illisible = NaN.
function parseAmount(raw) {
  const text = String(raw ?? '').trim()
  if (text === '') return 0
  const cleaned = text.replace(/\s/g, '').replace(',', '.').replace(/[^0-9.\-]/g, '')
  const n = cleaned === '' ? NaN : Number(cleaned)
  return Number.isFinite(n) ? n : NaN
}

const gainAmountParsed = computed(() => parseAmount(form.gainAmount))

const gainPreview = computed(() => {
  const amount = gainAmountParsed.value
  if (!(amount > 0)) return ''
  const pretty = amount.toLocaleString('fr-FR', { maximumFractionDigits: 2 })
  return `${pretty} ${form.gainUnit} ${form.gainLabel} par ${PERIOD_LABELS[form.gainPeriod]}`.replace(/\s+/g, ' ').trim()
})

// Unité ou libellé renseignés mais pas de montant : la tuile n'affichera rien, on le dit.
const gainMissingAmount = computed(
  () => String(form.gainAmount).trim() === '' && (form.gainUnit.trim() !== '' || form.gainLabel.trim() !== ''),
)

function pickUnit(u) {
  form.gainUnit = u.unit
  if (!form.gainLabel.trim() || UNITS.some((x) => x.label === form.gainLabel)) form.gainLabel = u.label
}
const error = ref('')
const titleInput = ref(null)

function submit() {
  error.value = ''
  if (!form.title.trim()) {
    error.value = 'Il faut un titre, même tout petit 🐣'
    return
  }
  const iso = fromLocalInput(form.startAt)
  if (!iso) {
    error.value = 'Cette date a un goût bizarre 🤔'
    return
  }
  const gainAmount = gainAmountParsed.value
  if (Number.isNaN(gainAmount)) {
    error.value = 'Le montant du gain doit être un nombre, par exemple 13 ou 13,5 🧮'
    return
  }
  if (gainAmount < 0) {
    error.value = 'Un gain négatif, ce serait une perte 😅'
    return
  }
  const equivValue = parseAmount(form.equivValue)
  if (Number.isNaN(equivValue) || equivValue < 0) {
    error.value = "La valeur de l'équivalence doit être un nombre, par exemple 12 🍕"
    return
  }
  emit('save', {
    emoji: form.emoji.trim() || '🍦',
    title: form.title.trim(),
    phrase: form.phrase.trim(),
    color: form.color,
    startAt: iso,
    gainAmount,
    gainUnit: form.gainUnit.trim(),
    gainPeriod: form.gainPeriod,
    gainLabel: form.gainLabel.trim(),
    equivUnit: form.equivUnit.trim(),
    equivValue,
  })
}

function setNow() {
  form.startAt = toLocalInput()
}

function onKey(e) {
  if (e.key === 'Escape') emit('close')
}
onMounted(() => {
  window.addEventListener('keydown', onKey)
  titleInput.value?.focus()
})
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <form class="card modal" @submit.prevent="submit">
      <h2 class="modal-title">{{ counter ? 'Retoucher le compteur ✏️' : 'Un nouveau compteur 🍦' }}</h2>

      <div class="field">
        <label for="f-emoji">Emoji</label>
        <div class="emoji-row">
          <input id="f-emoji" v-model="form.emoji" class="emoji-input" maxlength="8" />
          <div class="emoji-picks">
            <button
              v-for="e in EMOJIS"
              :key="e"
              type="button"
              class="emoji-pick"
              :class="{ active: form.emoji === e }"
              @click="form.emoji = e"
            >
              {{ e }}
            </button>
          </div>
        </div>
      </div>

      <div class="field">
        <label for="f-title">Titre</label>
        <input id="f-title" ref="titleInput" v-model="form.title" maxlength="80" placeholder="Sans café ☕" required />
      </div>

      <div class="field">
        <label for="f-phrase">Phrase (optionnelle)</label>
        <textarea id="f-phrase" v-model="form.phrase" maxlength="280" rows="2" placeholder="depuis que j'ai décidé d'être une personne raisonnable"></textarea>
      </div>

      <div class="field">
        <label for="f-start">Depuis le</label>
        <div class="date-row">
          <input id="f-start" v-model="form.startAt" type="datetime-local" required />
          <button type="button" class="btn btn-sm" @click="setNow">Maintenant ⏱️</button>
        </div>
      </div>

      <fieldset class="gain-fields">
        <legend>Et on y gagne quoi ? (optionnel)</legend>
        <div class="gain-grid">
          <div class="field">
            <label for="f-gain">Montant</label>
            <input id="f-gain" v-model="form.gainAmount" type="text" inputmode="decimal" placeholder="13 ou 13,5" />
          </div>
          <div class="field">
            <label for="f-unit">Unité</label>
            <input id="f-unit" v-model="form.gainUnit" maxlength="16" placeholder="€" />
          </div>
          <div class="field">
            <label for="f-period">Par</label>
            <select id="f-period" v-model="form.gainPeriod">
              <option value="hour">heure</option>
              <option value="day">jour</option>
              <option value="week">semaine</option>
              <option value="month">mois</option>
            </select>
          </div>
        </div>
        <div class="unit-picks">
          <button
            v-for="u in UNITS"
            :key="u.unit"
            type="button"
            class="unit-pick"
            :class="{ active: form.gainUnit === u.unit }"
            @click="pickUnit(u)"
          >
            {{ u.unit }}
          </button>
        </div>
        <div class="field">
          <label for="f-gain-label">Libellé</label>
          <input id="f-gain-label" v-model="form.gainLabel" maxlength="40" placeholder="économisés, évitées, non fumées…" />
        </div>
        <div class="gain-grid equiv-grid">
          <div class="field">
            <label for="f-equiv-unit">Équivalence rigolote</label>
            <input id="f-equiv-unit" v-model="form.equivUnit" maxlength="30" placeholder="pizzas 🍕" />
          </div>
          <div class="field">
            <label for="f-equiv-value">1 = combien de {{ form.gainUnit.trim() || 'unités' }} ?</label>
            <input id="f-equiv-value" v-model="form.equivValue" type="text" inputmode="decimal" placeholder="12" />
          </div>
        </div>
        <p v-if="gainPreview" class="hint">Ça fera : <b>{{ gainPreview }}</b> 🎉<span v-if="!form.equivUnit.trim()"> Sans équivalence, on pioche dans la liste maison (cafés, pizzas, Big Mac…).</span></p>
        <p v-else-if="gainMissingAmount" class="hint hint-warn">Sans montant, rien ne s'affichera sur la tuile : ajoute un chiffre ci-dessus 👆</p>
      </fieldset>

      <div class="field">
        <label>Couleur</label>
        <div class="swatches">
          <button
            v-for="c in COLORS"
            :key="c"
            type="button"
            class="swatch"
            :class="[`c-${c}`, { active: form.color === c }]"
            :title="c"
            @click="form.color = c"
          ></button>
        </div>
      </div>

      <p v-if="error" class="error" role="alert">{{ error }}</p>

      <div class="modal-actions">
        <button type="button" class="btn btn-ghost" @click="emit('close')">Annuler</button>
        <button type="submit" class="btn btn-primary">{{ counter ? 'Enregistrer ✨' : 'Lancer le compteur 🚀' }}</button>
      </div>
    </form>
  </div>
</template>
