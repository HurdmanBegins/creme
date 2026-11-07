import { formatNumber } from './time'

// Équivalences rigolotes intégrées, par famille d'unité. value = combien d'unités du gain pour 1 article.
const MONEY = ['€', '$', '£', 'chf', 'eur', 'euro', 'euros', 'usd']
const BUILTIN = {
  money: [
    { one: 'café', many: 'cafés', emoji: '☕', value: 2.5 },
    { one: 'croissant', many: 'croissants', emoji: '🥐', value: 1.3 },
    { one: 'pizza', many: 'pizzas', emoji: '🍕', value: 12 },
    { one: 'place de ciné', many: 'places de ciné', emoji: '🎬', value: 11 },
    { one: 'jeu vidéo', many: 'jeux vidéo', emoji: '🎮', value: 60 },
    { one: 'week-end à la mer', many: 'week-ends à la mer', emoji: '🏖️', value: 300 },
    { one: 'console', many: 'consoles', emoji: '🕹️', value: 500 },
    { one: 'voyage au Japon', many: 'voyages au Japon', emoji: '🗼', value: 2500 },
  ],
  kcal: [
    { one: 'Big Mac', many: 'Big Mac', emoji: '🍔', value: 500 },
    { one: 'pizza', many: 'pizzas', emoji: '🍕', value: 800 },
    { one: 'raclette', many: 'raclettes', emoji: '🧀', value: 1500 },
    { one: 'marathon', many: 'marathons', emoji: '🏃', value: 2600 },
  ],
  cigarettes: [
    { one: 'paquet', many: 'paquets', emoji: '🚬', value: 20 },
    { one: 'cartouche', many: 'cartouches', emoji: '📦', value: 200 },
  ],
  hours: [
    { one: 'film', many: 'films', emoji: '🎬', value: 2 },
    { one: 'nuit de sommeil', many: 'nuits de sommeil', emoji: '😴', value: 8 },
    { one: 'saison de série', many: 'saisons de série', emoji: '📺', value: 10 },
  ],
  drinks: [
    { one: 'pack', many: 'packs', emoji: '📦', value: 6 },
    { one: 'fût', many: 'fûts', emoji: '🛢️', value: 60 },
  ],
}

function family(unit) {
  const u = (unit || '').trim().toLowerCase()
  if (MONEY.includes(u)) return 'money'
  if (u.includes('cal')) return 'kcal'
  if (u.includes('cigarette') || u.includes('clope')) return 'cigarettes'
  if (u === 'h' || u.includes('heure')) return 'hours'
  if (u.includes('🍺') || u.includes('bière') || u.includes('verre')) return 'drinks'
  return null
}

function phrase(count, one, many, emoji) {
  const decimals = count < 10 ? 1 : 0
  const name = count < 2 ? one : many
  return `≈ ${formatNumber(count, decimals)} ${name}${emoji ? ' ' + emoji : ''}`
}

// Liste des équivalences applicables : la personnalisée d'abord, puis les intégrées
// dont le compte est parlant (au moins 0,5).
export function equivalences(counter, gainValue) {
  const out = []
  if (counter.equivUnit && Number(counter.equivValue) > 0) {
    const count = gainValue / Number(counter.equivValue)
    out.push(phrase(count, counter.equivUnit, counter.equivUnit, ''))
  }
  const list = BUILTIN[family(counter.gainUnit)] || []
  const fitting = list.filter((e) => gainValue / e.value >= 0.5)
  for (const e of fitting) out.push(phrase(gainValue / e.value, e.one, e.many, e.emoji))
  return out
}

// Une seule équivalence à la fois, qui tourne toutes les 8 secondes.
export function pickEquivalence(list, nowMs) {
  if (!list.length) return ''
  return list[Math.floor(nowMs / 8000) % list.length]
}
