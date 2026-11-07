import { ref, computed } from 'vue'

// Thème : « auto » suit le système, sinon « light » / « dark ». Mémorisé dans le navigateur.
const KEY = 'creme-theme'
const media = window.matchMedia?.('(prefers-color-scheme: dark)')
const theme = ref(read())
const systemDark = ref(!!media?.matches)
media?.addEventListener?.('change', (e) => (systemDark.value = e.matches))

function read() {
  try {
    const v = localStorage.getItem(KEY)
    return v === 'light' || v === 'dark' ? v : 'auto'
  } catch {
    return 'auto'
  }
}

export const isDark = computed(() => (theme.value === 'auto' ? systemDark.value : theme.value === 'dark'))

export function applyTheme() {
  document.documentElement.dataset.theme = isDark.value ? 'dark' : 'light'
  const meta = document.querySelector('meta[name=theme-color]')
  if (meta) meta.content = isDark.value ? '#241a2b' : '#ffd6e8'
}

export function toggleTheme() {
  theme.value = isDark.value ? 'light' : 'dark'
  try {
    localStorage.setItem(KEY, theme.value)
  } catch {
    /* pas grave */
  }
  applyTheme()
}

export function useTheme() {
  return { theme, isDark, toggleTheme, applyTheme }
}
