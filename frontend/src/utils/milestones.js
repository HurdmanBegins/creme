import { HOUR, DAY } from './time'

// Paliers de durée, du premier au plus glorieux.
export const DURATION_MILESTONES = [
  { key: 'h1', ms: HOUR, label: '1 heure', phrase: 'Une heure ! Ça commence 🐣' },
  { key: 'd1', ms: DAY, label: '24 heures', phrase: '24 h ! Une journée entière, respect 🌞' },
  { key: 'd3', ms: 3 * DAY, label: '3 jours', phrase: '3 jours ! Le plus dur est fait (on dit ça) 💪' },
  { key: 'w1', ms: 7 * DAY, label: '1 semaine', phrase: 'Une semaine ! Chantilly bien montée 🍦' },
  { key: 'w2', ms: 14 * DAY, label: '2 semaines', phrase: 'Deux semaines ! Le gâteau prend forme 🎂' },
  { key: 'm1', ms: 30 * DAY, label: '1 mois', phrase: "Un mois ! C'est officiellement une habitude 🏅" },
  { key: 'd100', ms: 100 * DAY, label: '100 jours', phrase: '100 jours ! Bienvenue au club très fermé 💯' },
  { key: 'm6', ms: 182 * DAY, label: '6 mois', phrase: 'Six mois ! On ne te reconnaît plus 🦋' },
  { key: 'y1', ms: 365 * DAY, label: '1 an', phrase: 'UN AN ! Légende vivante 👑' },
  { key: 'y2', ms: 730 * DAY, label: '2 ans', phrase: 'Deux ans ! Immortel·le 🧓' },
  { key: 'y5', ms: 5 * 365 * DAY, label: '5 ans', phrase: 'Cinq ans ! On devrait te faire une statue 🗿' },
  { key: 'y10', ms: 10 * 365 * DAY, label: '10 ans', phrase: 'Dix ans ! Tu es devenu·e une légende urbaine 🏛️' },
]

// Paliers de gain : 10, 25, 50, 100, 250, 500, 1 000, 2 500…
const GAIN_STEPS = [1, 2.5, 5]
export function gainMilestoneSteps(max = 1e9) {
  const out = []
  for (let p = 10; p <= max; p *= 10) for (const s of GAIN_STEPS) if (p * s <= max) out.push(p * s)
  return out
}
const GAIN_MILESTONES = gainMilestoneSteps()

export function durationMilestone(ms) {
  let reachedIndex = -1
  for (let i = 0; i < DURATION_MILESTONES.length; i++) if (ms >= DURATION_MILESTONES[i].ms) reachedIndex = i
  const reached = reachedIndex >= 0 ? DURATION_MILESTONES[reachedIndex] : null
  const next = DURATION_MILESTONES[reachedIndex + 1] || null
  const from = reached ? reached.ms : 0
  const progress = next ? Math.min(1, Math.max(0, (ms - from) / (next.ms - from))) : 1
  return { reachedIndex, reached, next, progress, remainingMs: next ? next.ms - ms : 0 }
}

export function gainMilestone(value) {
  let reachedIndex = -1
  for (let i = 0; i < GAIN_MILESTONES.length; i++) if (value >= GAIN_MILESTONES[i]) reachedIndex = i
  return {
    reachedIndex,
    reached: reachedIndex >= 0 ? GAIN_MILESTONES[reachedIndex] : null,
    next: GAIN_MILESTONES[reachedIndex + 1] || null,
  }
}

// Mémoire des paliers déjà fêtés (par compteur) pour ne fêter que les nouveaux.
const STORAGE_KEY = 'creme-milestones'
function readSeen() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}')
  } catch {
    return {}
  }
}
export function getSeen(id) {
  return readSeen()[id] || null
}
export function setSeen(id, value) {
  try {
    const all = readSeen()
    all[id] = value
    localStorage.setItem(STORAGE_KEY, JSON.stringify(all))
  } catch {
    /* stockage indisponible : tant pis, pas de confettis mémorisés */
  }
}
export function forgetSeen(id) {
  try {
    const all = readSeen()
    delete all[id]
    localStorage.setItem(STORAGE_KEY, JSON.stringify(all))
  } catch {
    /* idem */
  }
}
