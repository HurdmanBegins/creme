import { ref } from 'vue'

// Horloge partagée par toutes les cartes : un seul setInterval pour tout le monde.
// `offset` corrige la dérive entre l'horloge du navigateur et celle du serveur.
let offset = 0
export const now = ref(Date.now())

export function setServerTime(iso) {
  const server = new Date(iso).getTime()
  if (!Number.isNaN(server)) offset = server - Date.now()
}

setInterval(() => {
  now.value = Date.now() + offset
}, 250)
