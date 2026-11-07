import { splitDuration, pad } from './time'

const PALETTE = {
  cream: ['#fff3dc', '#ffe2b8'],
  pink: ['#ffe0ec', '#ffbfd8'],
  lavender: ['#ece2ff', '#d2bdff'],
  mint: ['#dcf8ec', '#b6ecd4'],
  sky: ['#dff0ff', '#b9dcff'],
  peach: ['#ffe8d3', '#ffcda5'],
  lemon: ['#fff7c7', '#fff0a0'],
}

function roundRect(ctx, x, y, w, h, r) {
  ctx.beginPath()
  ctx.moveTo(x + r, y)
  ctx.arcTo(x + w, y, x + w, y + h, r)
  ctx.arcTo(x + w, y + h, x, y + h, r)
  ctx.arcTo(x, y + h, x, y, r)
  ctx.arcTo(x, y, x + w, y, r)
  ctx.closePath()
}

function wrap(ctx, text, maxWidth, maxLines) {
  const words = text.split(/\s+/)
  const lines = []
  let line = ''
  for (const w of words) {
    const test = line ? `${line} ${w}` : w
    if (ctx.measureText(test).width > maxWidth && line) {
      lines.push(line)
      line = w
    } else line = test
  }
  if (line) lines.push(line)
  if (lines.length > maxLines) {
    lines.length = maxLines
    lines[maxLines - 1] = lines[maxLines - 1].replace(/\s*\S*$/, '') + '…'
  }
  return lines
}

// Dessine la carte 1080×1080 d'un compteur et renvoie un Blob PNG.
export async function renderCardPng(counter, { elapsedMs, gainText, equivText, mood }) {
  await document.fonts?.ready
  const size = 1080
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')
  const [c1, c2] = PALETTE[counter.color] || PALETTE.cream

  // fond + vermicelles
  const bg = ctx.createLinearGradient(0, 0, size, size)
  bg.addColorStop(0, c1)
  bg.addColorStop(1, c2)
  ctx.fillStyle = bg
  ctx.fillRect(0, 0, size, size)
  const colors = ['#ff8fb8', '#c78bff', '#7fd8b8', '#7fc4ff', '#ffb86b', '#ffe066']
  for (let i = 0; i < 40; i++) {
    ctx.save()
    ctx.translate(((i * 263) % size) + 20, ((i * 397) % size) + 20)
    ctx.rotate(((i * 47) % 180) * (Math.PI / 180))
    ctx.fillStyle = colors[i % colors.length]
    ctx.globalAlpha = 0.5
    roundRect(ctx, -6, -16, 12, 32, 6)
    ctx.fill()
    ctx.restore()
  }

  // carte
  ctx.save()
  ctx.shadowColor = 'rgba(120,70,90,0.25)'
  ctx.shadowBlur = 40
  ctx.shadowOffsetY = 16
  ctx.fillStyle = '#fffdf9'
  roundRect(ctx, 90, 110, size - 180, size - 220, 56)
  ctx.fill()
  ctx.restore()

  const ink = '#4a2f35'
  const soft = '#8a6a72'
  const display = '"Fredoka", "Nunito", system-ui, sans-serif'
  const body = '"Nunito", system-ui, sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'

  // emoji + titre + phrase
  ctx.font = `150px ${body}`
  ctx.fillText(counter.emoji || '🍦', size / 2, 250)
  ctx.fillStyle = ink
  ctx.font = `700 64px ${display}`
  const titleLines = wrap(ctx, counter.title, size - 300, 2)
  titleLines.forEach((l, i) => ctx.fillText(l, size / 2, 375 + i * 72))
  let y = 375 + titleLines.length * 72
  if (counter.phrase) {
    ctx.fillStyle = soft
    ctx.font = `italic 600 34px ${body}`
    const lines = wrap(ctx, counter.phrase, size - 300, 2)
    lines.forEach((l, i) => ctx.fillText(l, size / 2, y + i * 42))
    y += lines.length * 42 + 10
  }

  // durée
  y += 30
  ctx.fillStyle = soft
  ctx.font = `600 28px ${display}`
  ctx.fillText(elapsedMs < 0 ? 'ÇA COMMENCE DANS' : 'DEPUIS', size / 2, y)
  y += 70
  const p = splitDuration(Math.abs(elapsedMs))
  const units = []
  if (p.years) units.push([String(p.years), p.years > 1 ? 'ans' : 'an'])
  units.push([String(p.days), p.days > 1 ? 'jours' : 'jour'], [pad(p.hours), 'h'], [pad(p.minutes), 'min'], [pad(p.seconds), 's'])
  const pillW = Math.min(170, (size - 260 - (units.length - 1) * 16) / units.length)
  const totalW = units.length * pillW + (units.length - 1) * 16
  let x = (size - totalW) / 2
  for (const [val, lab] of units) {
    ctx.fillStyle = c1
    roundRect(ctx, x, y - 50, pillW, 110, 28)
    ctx.fill()
    ctx.fillStyle = lab === 's' ? '#ff8fb8' : ink
    ctx.font = `700 54px ${display}`
    ctx.fillText(val, x + pillW / 2, y - 6)
    ctx.fillStyle = soft
    ctx.font = `800 20px ${body}`
    ctx.fillText(lab.toUpperCase(), x + pillW / 2, y + 36)
    x += pillW + 16
  }
  y += 120

  // gain + équivalence
  if (gainText) {
    ctx.fillStyle = '#ff8fb8'
    ctx.font = `700 60px ${display}`
    ctx.fillText(gainText, size / 2, y)
    y += 60
    if (equivText) {
      ctx.fillStyle = soft
      ctx.font = `700 32px ${body}`
      ctx.fillText(equivText, size / 2, y)
      y += 50
    }
  }

  // humeur + pied
  ctx.fillStyle = soft
  ctx.font = `700 30px ${body}`
  const moodLines = wrap(ctx, mood, size - 300, 2)
  moodLines.forEach((l, i) => ctx.fillText(l, size / 2, Math.max(y + 20, 820) + i * 38))
  ctx.fillStyle = ink
  ctx.font = `700 34px ${display}`
  ctx.fillText(`Crême 2.0  ·  ${location.host}`, size / 2, size - 150)

  return new Promise((resolve) => canvas.toBlob(resolve, 'image/png'))
}

// Partage natif si possible (mobile), sinon téléchargement.
export async function shareOrDownload(blob, filename, title) {
  const file = new File([blob], filename, { type: 'image/png' })
  if (navigator.canShare && navigator.canShare({ files: [file] })) {
    try {
      await navigator.share({ files: [file], title })
      return 'shared'
    } catch (e) {
      if (e?.name === 'AbortError') return 'cancelled'
    }
  }
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  setTimeout(() => URL.revokeObjectURL(url), 2000)
  return 'downloaded'
}
