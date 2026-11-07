<script setup>
// La mascotte : une coupe de chantilly avec une cerise, et des yeux qui changent d'humeur.
defineProps({
  mood: { type: String, default: 'happy' }, // happy | shy | sleepy
  night: { type: Boolean, default: false }, // bonnet de nuit en mode sombre
})
</script>

<template>
  <svg class="mascot" viewBox="0 0 120 120" role="img" aria-label="Mascotte Crême 2.0">
    <!-- coupe -->
    <path d="M30 86 Q60 100 90 86 L84 108 Q60 116 36 108 Z" fill="#ffb3c6" />
    <ellipse cx="60" cy="86" rx="31" ry="7" fill="#ff8fb8" />
    <!-- chantilly -->
    <ellipse cx="60" cy="78" rx="38" ry="16" fill="#fff8ee" stroke="#f7e3cf" stroke-width="2" />
    <ellipse cx="60" cy="62" rx="30" ry="15" fill="#fffbf4" stroke="#f7e3cf" stroke-width="2" />
    <ellipse cx="60" cy="48" rx="20" ry="12" fill="#ffffff" stroke="#f7e3cf" stroke-width="2" />
    <!-- cerise, ou bonnet de nuit -->
    <g v-if="night">
      <path d="M42 46 Q52 10 84 18 Q74 30 80 46 Z" fill="#c78bff" stroke="#a56ee6" stroke-width="2" stroke-linejoin="round" />
      <circle cx="85" cy="17" r="5.5" fill="#fff" stroke="#e8d9ff" stroke-width="1.5" />
      <path d="M40 47 Q60 40 82 47" stroke="#fff" stroke-width="6" fill="none" stroke-linecap="round" />
      <text x="96" y="30" font-size="11" font-weight="700" fill="#c9a9b6" font-family="Fredoka, sans-serif">z</text>
    </g>
    <g v-else>
      <circle cx="61" cy="34" r="8" fill="#ff5c8a" />
      <circle cx="58" cy="31" r="2.2" fill="#fff" opacity=".8" />
      <path d="M62 26 q4 -12 14 -9" stroke="#6bbf59" stroke-width="2.5" fill="none" stroke-linecap="round" />
    </g>

    <!-- visage -->
    <g v-if="mood === 'sleepy'">
      <path d="M46 64 q4 4 8 0" stroke="#4a2f35" stroke-width="2.5" fill="none" stroke-linecap="round" />
      <path d="M66 64 q4 4 8 0" stroke="#4a2f35" stroke-width="2.5" fill="none" stroke-linecap="round" />
      <text x="84" y="46" font-size="14" font-weight="700" fill="#8a6a72" font-family="Fredoka, sans-serif">z</text>
      <text x="92" y="36" font-size="10" font-weight="700" fill="#8a6a72" font-family="Fredoka, sans-serif">z</text>
    </g>
    <g v-else-if="mood === 'shy'">
      <path d="M45 66 l5 -5 l5 5" stroke="#4a2f35" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M65 66 l5 -5 l5 5" stroke="#4a2f35" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round" />
    </g>
    <g v-else>
      <circle cx="50" cy="64" r="3.2" fill="#4a2f35" />
      <circle cx="70" cy="64" r="3.2" fill="#4a2f35" />
      <circle cx="51.2" cy="62.8" r="1" fill="#fff" />
      <circle cx="71.2" cy="62.8" r="1" fill="#fff" />
    </g>
    <!-- joues -->
    <ellipse cx="42" cy="71" rx="4.5" ry="2.6" fill="#ffb3c6" :opacity="mood === 'shy' ? 1 : 0.8" />
    <ellipse cx="78" cy="71" rx="4.5" ry="2.6" fill="#ffb3c6" :opacity="mood === 'shy' ? 1 : 0.8" />
    <!-- bouche -->
    <path v-if="mood === 'sleepy'" d="M57 73 q3 2 6 0" stroke="#4a2f35" stroke-width="2" fill="none" stroke-linecap="round" />
    <path v-else d="M54 71 q3 4 6 0 q3 4 6 0" stroke="#4a2f35" stroke-width="2" fill="none" stroke-linecap="round" />
  </svg>
</template>
