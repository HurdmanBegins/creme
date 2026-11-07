export const SECOND = 1000
export const MINUTE = 60 * SECOND
export const HOUR = 60 * MINUTE
export const DAY = 24 * HOUR
export const YEAR = 365 * DAY

// Temps écoulé en ms. Si le compteur est en pause, il reste figé sur pausedAt.
export function elapsedMs(counter, nowMs) {
  const end = counter.paused && counter.pausedAt ? new Date(counter.pausedAt).getTime() : nowMs
  return end - new Date(counter.startAt).getTime() - (counter.pausedSeconds || 0) * SECOND
}

export function splitDuration(ms) {
  let rest = Math.max(0, Math.floor(ms / SECOND))
  const years = Math.floor(rest / (YEAR / SECOND))
  rest -= years * (YEAR / SECOND)
  const days = Math.floor(rest / (DAY / SECOND))
  rest -= days * (DAY / SECOND)
  const hours = Math.floor(rest / 3600)
  rest -= hours * 3600
  const minutes = Math.floor(rest / 60)
  const seconds = rest - minutes * 60
  return { years, days, hours, minutes, seconds }
}

export const pad = (n) => String(n).padStart(2, '0')

const dateFmt = new Intl.DateTimeFormat('fr-FR', { dateStyle: 'long', timeStyle: 'short' })
export function formatDateTime(iso) {
  return dateFmt.format(new Date(iso))
}

// Petit commentaire humoristique selon la durée.
export function moodFor(ms, paused) {
  if (paused) return 'En pause. Le temps fait la sieste 💤'
  if (ms < 0) return 'Pas encore commencé, on fouette la chantilly ⏳'
  if (ms < MINUTE) return 'Tout frais, encore tiède 🥐'
  if (ms < HOUR) return 'Ça vient à peine de commencer 🐣'
  if (ms < DAY) return 'Quelques heures, ça passe crème 🍦'
  if (ms < 7 * DAY) return 'Déjà plusieurs jours, hein 👀'
  if (ms < 30 * DAY) return 'Des semaines… ça commence à faire long 🫠'
  if (ms < YEAR) return 'Des mois ! On fait un gâteau ? 🎂'
  return 'Une éternité. Respect. 🧓'
}

// ISO -> valeur pour <input type="datetime-local"> (heure locale, sans secondes)
export function toLocalInput(iso) {
  const d = iso ? new Date(iso) : new Date()
  const local = new Date(d.getTime() - d.getTimezoneOffset() * MINUTE)
  return local.toISOString().slice(0, 16)
}

// valeur datetime-local -> ISO UTC
export function fromLocalInput(value) {
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? null : d.toISOString()
}

// ---------- gains ----------
const PERIOD_MS = { hour: HOUR, day: DAY, week: 7 * DAY, month: 30.4375 * DAY }
export const PERIOD_LABELS = { hour: 'heure', day: 'jour', week: 'semaine', month: 'mois' }
const MONEY = ['€', '$', '£', 'chf', 'eur', 'euro', 'euros', 'usd']

export function hasGain(counter) {
  return Number(counter.gainAmount) > 0
}

// Valeur accumulée : montant × temps écoulé / période. Figée en pause (via ms), nulle avant le départ.
export function gainValue(counter, ms) {
  const amount = Number(counter.gainAmount) || 0
  if (amount <= 0 || ms <= 0) return 0
  return (amount * ms) / (PERIOD_MS[counter.gainPeriod] || DAY)
}

export function gainDecimals(counter) {
  const unit = (counter.gainUnit || '').trim().toLowerCase()
  if (MONEY.includes(unit)) return 2
  return Number(counter.gainAmount) < 100 ? 1 : 0
}

export function formatNumber(value, decimals) {
  return new Intl.NumberFormat('fr-FR', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(value)
}

export function gainEmoji(unit) {
  const u = (unit || '').trim().toLowerCase()
  if (!u) return '✨'
  if (MONEY.includes(u)) return '💰'
  if (u.includes('cal')) return '🔥'
  if (u.includes('cigarette') || u.includes('clope')) return '🚭'
  if (u === 'h' || u.includes('heure') || u.includes('min')) return '⏰'
  if (u.includes('km') || u.includes('pas')) return '🏃'
  if (/\p{Extended_Pictographic}/u.test(u)) return '' // l'unité est déjà un emoji
  return '✨'
}

// Durée courte et lisible : « 23 j 4 h », « 3 h 12 min », « 45 s ».
export function formatShort(ms) {
  const p = splitDuration(Math.max(0, ms))
  const days = p.years * 365 + p.days
  if (days >= 1) return p.hours ? `${days} j ${p.hours} h` : `${days} j`
  if (p.hours >= 1) return p.minutes ? `${p.hours} h ${pad(p.minutes)} min` : `${p.hours} h`
  if (p.minutes >= 1) return `${p.minutes} min`
  return `${p.seconds} s`
}
