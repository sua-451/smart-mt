<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'
  import { statusClass } from '../utils.js'

  let { currentUser } = $props()

  let assets = $state([])
  let loading = $state(true)
  let flashMsg = $state({ type: '', text: '' })

  async function loadAssets() {
    loading = true
    try {
      assets = await api.assets.list()
    } catch (err) {
      flashMsg = { type: 'error', text: err.message || 'Gagal mengambil data aset' }
    } finally {
      loading = false
    }
  }

  async function handleRetire(id) {
    if (!confirm('Tandai aset ini sebagai afkir? Riwayat tetap tersimpan.')) {
      return
    }
    try {
      const res = await api.assets.retire(id)
      flashMsg = { type: 'success', text: res.message || 'Aset berhasil ditandai afkir.' }
      loadAssets()
    } catch (err) {
      flashMsg = { type: 'error', text: err.message || 'Gagal menandai aset afkir' }
    }
  }

  onMount(() => {
    loadAssets()
  })
</script>

<div class="page-heading">
  <div>
    <span class="eyebrow">DATA ASET</span>
    <h1>{currentUser?.role === 'Manager' ? 'Master aset' : 'Aset mesin'}</h1>
    <p>
      {currentUser?.role === 'Manager'
        ? 'Kelola aset dan buka kartu mesin untuk melihat informasi serta riwayatnya.'
        : 'Pilih kartu mesin untuk melihat informasi lengkap dan riwayat maintenance.'}
    </p>
  </div>
  {#if currentUser?.role === 'Manager'}
    <a class="button primary" href="#/assets/new">
      <span>＋</span> Tambah aset
    </a>
  {/if}
</div>

{#if flashMsg.text}
  <div class="flash-stack">
    <div class="flash {flashMsg.type}">{flashMsg.text}</div>
  </div>
{/if}

{#if loading}
  <section class="panel">
    <div class="empty-state">
      <span>⬡</span>
      <strong>Memuat daftar aset...</strong>
    </div>
  </section>
{:else if assets && assets.length > 0}
  <div class="asset-card-grid">
    {#each assets as asset}
      <article class="asset-card">
        <a class="asset-card-main" href="#/assets/{asset.id_asset}">
          <div class="asset-card-image">
            {#if asset.first_photo}
              <img
                src={api.uploadUrl(asset.first_photo)}
                alt="Foto {asset.nama_mesin}"
                loading="lazy"
              />
            {:else}
              <div class="asset-card-placeholder">
                <span>⬡</span>
                <span>Belum ada foto mesin</span>
              </div>
            {/if}
            <span class="status-pill {statusClass(asset.status)}">{asset.status}</span>
            {#if asset.photo_count > 1}
              <span class="asset-card-photo-count">▧ {asset.photo_count} foto</span>
            {/if}
          </div>
          <div class="asset-card-info">
            <span class="code-chip">
              {asset.kode_asset}{asset.serial_number ? ` · SN ${asset.serial_number}` : ''}
            </span>
            <h2>{asset.nama_mesin}</h2>
            <p>
              {asset.merk || 'Merek tidak dicatat'}{asset.tipe_model ? ` · ${asset.tipe_model}` : ''}<br />
              {asset.lokasi || 'Lokasi tidak dicatat'}
            </p>
            <div class="asset-card-meta">
              <span>◷</span> {asset.work_order_count || 0} riwayat maintenance
            </div>
            <span class="asset-card-link">Lihat detail &amp; riwayat <span>→</span></span>
          </div>
        </a>
        {#if currentUser?.role === 'Manager'}
          <div class="asset-card-actions">
            <a class="button small secondary" href="#/assets/{asset.id_asset}/edit">
              Edit aset
            </a>
            {#if asset.status !== 'Afkir'}
              <button
                class="button small danger-outline"
                type="button"
                onclick={() => handleRetire(asset.id_asset)}
              >
                Afkir
              </button>
            {/if}
          </div>
        {/if}
      </article>
    {/each}
  </div>
{:else}
  <section class="panel">
    <div class="empty-state">
      <span>⬡</span>
      <strong>Belum ada aset</strong>
      <p>
        {currentUser?.role === 'Manager'
          ? 'Tambahkan aset pertama untuk memulai jadwal maintenance.'
          : 'Aset mesin akan muncul di sini setelah didaftarkan manager.'}
      </p>
      {#if currentUser?.role === 'Manager'}
        <a class="button primary" href="#/assets/new">Tambah aset</a>
      {/if}
    </div>
  </section>
{/if}

