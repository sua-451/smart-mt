const BASE = 'http://localhost:8080/api'

export function getToken() {
  return localStorage.getItem('sm_token')
}

export function getUser() {
  try {
    return JSON.parse(localStorage.getItem('sm_user') || 'null')
  } catch {
    return null
  }
}

export function setAuth(token, user) {
  localStorage.setItem('sm_token', token)
  localStorage.setItem('sm_user', JSON.stringify(user))
}

export function clearAuth() {
  localStorage.removeItem('sm_token')
  localStorage.removeItem('sm_user')
}

export async function request(method, path, body = null, isMultipart = false) {
  const headers = {}
  const token = getToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  if (!isMultipart && body) {
    headers['Content-Type'] = 'application/json'
  }

  const options = {
    method,
    headers,
  }

  if (body) {
    options.body = isMultipart ? body : JSON.stringify(body)
  }

  const res = await fetch(`${BASE}${path}`, options)

  if (res.status === 401) {
    clearAuth()
    window.location.hash = '#/login'
    throw new Error('Sesi berakhir, silakan masuk kembali.')
  }

  const contentType = res.headers.get('content-type') || ''
  let data
  if (contentType.includes('application/json')) {
    data = await res.json()
  } else {
    data = await res.text()
  }

  if (!res.ok) {
    const errorMsg = (data && data.error) || (typeof data === 'string' ? data : 'Terjadi kesalahan sistem')
    throw new Error(errorMsg)
  }

  return data
}

export const api = {
  login: (username, password) =>
    request('POST', '/auth/login', { username, password }),
  
  me: () =>
    request('GET', '/auth/me'),

  dashboard: () =>
    request('GET', '/dashboard'),

  assets: {
    list: () => request('GET', '/assets'),
    get: (id) => request('GET', `/assets/${id}`),
    create: (formData) => request('POST', '/assets', formData, true),
    update: (id, formData) => request('PUT', `/assets/${id}`, formData, true),
    retire: (id) => request('DELETE', `/assets/${id}`),
  },

  workOrders: {
    list: () => request('GET', '/work-orders'),
    get: (id) => request('GET', `/work-orders/${id}`),
    create: (data) => request('POST', '/work-orders', data),
    assign: (id, technicianId) =>
      request('POST', `/work-orders/${id}/assign`, { technician_id: Number(technicianId) }),
    review: (id, action, managerNote = '') =>
      request('POST', `/work-orders/${id}/review`, { action, manager_note: managerNote }),
    report: (id, formData) =>
      request('POST', `/work-orders/${id}/report`, formData, true),
  },

  users: {
    list: () => request('GET', '/users'),
    create: (data) => request('POST', '/users', data),
  },

  technicians: () =>
    request('GET', '/technicians'),

  uploadUrl: (filename) =>
    filename ? `http://localhost:8080/api/uploads/${filename}` : null,
}
