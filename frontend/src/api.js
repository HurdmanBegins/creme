const TOKEN_KEY = 'creme-token'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token) {
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

export class ApiError extends Error {
  constructor(status, message) {
    super(message)
    this.status = status
  }
}

async function request(method, path, body) {
  const headers = { Accept: 'application/json' }
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`
  if (body !== undefined) headers['Content-Type'] = 'application/json'

  let res
  try {
    res = await fetch(`/api${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError(0, 'Le serveur ne répond pas… il fait la sieste ? 😴')
  }

  if (res.status === 204) return null
  let data = null
  try {
    data = await res.json()
  } catch {
    /* corps vide ou non JSON */
  }
  if (!res.ok) throw new ApiError(res.status, data?.error || `Erreur ${res.status}`)
  return data
}

export const api = {
  login: (password) => request('POST', '/login', { password }),
  me: () => request('GET', '/me'),
  list: () => request('GET', '/counters'),
  create: (counter) => request('POST', '/counters', counter),
  update: (id, counter) => request('PUT', `/counters/${id}`, counter),
  remove: (id) => request('DELETE', `/counters/${id}`),
  pause: (id) => request('POST', `/counters/${id}/pause`),
  resume: (id) => request('POST', `/counters/${id}/resume`),
  reset: (id) => request('POST', `/counters/${id}/reset`),
  archive: (id) => request('POST', `/counters/${id}/archive`),
  unarchive: (id) => request('POST', `/counters/${id}/unarchive`),
}
