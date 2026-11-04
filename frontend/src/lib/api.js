const BASE_URL = 'http://localhost:8080/api';

export async function fetchAssets() {
  const res = await fetch(`${BASE_URL}/assets`);
  if (!res.ok) throw new Error('Gagal mengambil data aset');
  return res.json();
}

/** @param {any} data */
export async function createAsset(data) {
  const res = await fetch(`${BASE_URL}/assets`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  });
  return res.json();
}

/** @param {number|string} id */
export async function deleteAsset(id) {
  const res = await fetch(`${BASE_URL}/assets/${id}`, {
    method: 'DELETE'
  });
  return res.json();
}