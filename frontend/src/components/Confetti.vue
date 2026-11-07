<script setup>
import { ref } from 'vue'

// Pluie de confettis maison : burst() crée des pièces qui tombent puis disparaissent.
const COLORS = ['#ff8fb8', '#c78bff', '#7fd8b8', '#7fc4ff', '#ffb86b', '#ffe066', '#ffffff']
const pieces = ref([])
let seq = 0

function burst(count = 110) {
  const batch = Array.from({ length: count }, () => ({
    id: ++seq,
    style: {
      left: `${Math.random() * 100}%`,
      background: COLORS[Math.floor(Math.random() * COLORS.length)],
      animationDelay: `${Math.random() * 0.8}s`,
      animationDuration: `${2.2 + Math.random() * 1.4}s`,
      transform: `rotate(${Math.random() * 360}deg)`,
      width: `${8 + Math.random() * 8}px`,
      height: `${10 + Math.random() * 14}px`,
      borderRadius: Math.random() > 0.5 ? '50%' : '3px',
    },
  }))
  pieces.value.push(...batch)
  const ids = new Set(batch.map((p) => p.id))
  setTimeout(() => (pieces.value = pieces.value.filter((p) => !ids.has(p.id))), 4500)
}

defineExpose({ burst })
</script>

<template>
  <div class="confetti" aria-hidden="true">
    <span v-for="p in pieces" :key="p.id" class="confetti-piece" :style="p.style"></span>
  </div>
</template>
