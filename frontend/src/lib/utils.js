export function statusClass(status) {
  const map = {
    'Menunggu Penugasan': 'status-wait',
    'Ditugaskan': 'status-info',
    'Dikerjakan': 'status-progress',
    'Perlu Revisi': 'status-review',
    'Menunggu Approval': 'status-wait',
    'Terjadwal': 'status-wait',
    'Terlambat': 'status-review',
    'Terjadwal PM': 'status-wait',
    'Selesai': 'status-done',
    'Terlaksana': 'status-done',
    'Operasional': 'status-done',
    'Tidak Operasional': 'status-review',
    'Perbaikan': 'status-wait',
    'Afkir': 'status-muted',
    'Open': 'status-wait',
    'Proses': 'status-progress',
    'Batal': 'status-muted',
  }
  return map[status] || 'status-muted'
}

export function formatDate(val) {
  if (!val) return '—'
  const d = new Date(val)
  if (isNaN(d.getTime())) return val
  return d.toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
  })
}

export function today() {
  return new Date().toISOString().slice(0, 10)
}

