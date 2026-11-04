<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'
  import { statusClass } from '../utils.js'

  let { currentUser } = $props()

  let metrics = $state({
    assets: 0,
    active_orders: 0,
    pending_approval: 0,
    technicians: 0,
    completed: 0,
  })
  let orders = $state([])
  let loading = $state(true)
  let errorMsg = $state('')

  async function loadDashboard() {
    loading = true
    errorMsg = ''
    try {
      const data = await api.dashboard()
      if (data) {
        metrics = data.metrics || metrics
        orders = data.orders || []
      }
    } catch (err) {
      errorMsg = err.message || 'Gagal memuat dashboard'
    } finally {
      loading = false
    }
  }

  onMount(() => {
    loadDashboard()
  })

  let firstName = $derived(
    currentUser && currentUser.full_name
      ? currentUser.full_name.split(' ')[0]
      : 'User'
  )
</script>

{#if errorMsg}
  <div class="flash-stack">
    <div class="flash error">{errorMsg}</div>
  </div>
{/if}

<div class="page-heading">
  <div>
    <span class="eyebrow">OVERVIEW</span>
    <h1>Selamat datang, {firstName} 👋</h1>
    <p>
      {#if currentUser?.role === 'Manager'}
        Pantau kondisi aset dan aktivitas maintenance hari ini.
      {:else}
        Berikut daftar tugas maintenance yang perlu Anda kerjakan.
      {/if}
    </p>
  </div>
  {#if currentUser?.role === 'Manager'}
    <a class="button primary" href="#/work-orders">
      Kelola work order <span>→</span>
    </a>
  {/if}
</div>

{#if loading}
  <section class="panel">
    <div class="empty-state">
      <span>◷</span>
      <strong>Memuat data dashboard...</strong>
    </div>
  </section>
{:else if currentUser?.role === 'Manager'}
  <div class="stat-grid">
    <article class="stat-card">
      <div class="stat-top">
        <span>Aset aktif</span>
        <span class="stat-icon blue">⬡</span>
      </div>
      <strong>{metrics.assets}</strong>
      <small>Unit terdaftar dalam sistem</small>
    </article>
    <article class="stat-card">
      <div class="stat-top">
        <span>Work order aktif</span>
        <span class="stat-icon amber">▤</span>
      </div>
      <strong>{metrics.active_orders}</strong>
      <small>Perlu ditugaskan atau dikerjakan</small>
    </article>
    <article class="stat-card">
      <div class="stat-top">
        <span>Menunggu approval</span>
        <span class="stat-icon green">✓</span>
      </div>
      <strong>{metrics.pending_approval}</strong>
      <small>Laporan menunggu peninjauan</small>
    </article>
    <article class="stat-card">
      <div class="stat-top">
        <span>Teknisi</span>
        <span class="stat-icon violet">♙</span>
      </div>
      <strong>{metrics.technicians}</strong>
      <small>Akun teknisi aktif</small>
    </article>
  </div>

  <section class="panel">
    <div class="panel-heading">
      <div>
        <h2>Aktivitas work order</h2>
        <p>Jadwal maintenance dan laporan terbaru</p>
      </div>
      <a class="text-link" href="#/work-orders">Lihat semua <span>→</span></a>
    </div>

    {#if orders && orders.length > 0}
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Mesin</th>
              <th>Jenis</th>
              <th>Jadwal</th>
              <th>Teknisi</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {#each orders as order}
              <tr>
                <td>
                  <strong>{order.machine_name}</strong>
                  <small>{order.asset_id} · {order.location}</small>
                </td>
                <td>
                  {order.maintenance_type}
                  <small>Kerusakan {(order.severity || '').toLowerCase()}</small>
                </td>
                <td>{order.scheduled_date}</td>
                <td>{order.technician_name || 'Belum ditugaskan'}</td>
                <td>
                  <span class="status-pill {statusClass(order.status)}">
                    {order.status}
                  </span>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <div class="empty-state">
        <span>▤</span>
        <strong>Belum ada aktivitas</strong>
        <p>Tambahkan aset untuk mulai menjadwalkan maintenance.</p>
      </div>
    {/if}
  </section>
{:else}
  <!-- Teknisi View -->
  <div class="stat-grid technician-stats">
    <article class="stat-card">
      <div class="stat-top">
        <span>Tugas aktif</span>
        <span class="stat-icon blue">▤</span>
      </div>
      <strong>{metrics.active_orders}</strong>
      <small>Perlu dikerjakan atau direvisi</small>
    </article>
    <article class="stat-card">
      <div class="stat-top">
        <span>Menunggu approval</span>
        <span class="stat-icon amber">◷</span>
      </div>
      <strong>{metrics.pending_approval}</strong>
      <small>Laporan sudah dikirim</small>
    </article>
    <article class="stat-card">
      <div class="stat-top">
        <span>Tiket selesai</span>
        <span class="stat-icon green">✓</span>
      </div>
      <strong>{metrics.completed}</strong>
      <small>Maintenance disetujui manager</small>
    </article>
  </div>

  <div class="section-heading">
    <div>
      <h2>Tugas maintenance saya</h2>
      <p>Buka tugas untuk mencatat hasil pemeriksaan.</p>
    </div>
  </div>

  {#if orders && orders.length > 0}
    <div class="task-grid">
      {#each orders as order}
        <article class="task-card">
          <div class="task-card-top">
            <span class="status-pill {statusClass(order.status)}">{order.status}</span>
            <span class="task-date">{order.scheduled_date}</span>
          </div>
          <h3>{order.machine_name}</h3>
          <p class="task-meta">{order.asset_id} <span>·</span> {order.location}</p>
          <div class="task-details">
            <span>{order.maintenance_type}</span>
            <span>Tingkat {(order.severity || '').toLowerCase()}</span>
          </div>
          {#if order.status === 'Perlu Revisi'}
            <div class="revision-note">
              <strong>Catatan manager</strong>
              <p>{order.manager_note || 'Mohon tinjau kembali laporan.'}</p>
            </div>
          {/if}
          <a
            class="button primary full-button"
            href="#/work-orders/{order.ticket_id}/report"
          >
            {order.status === 'Perlu Revisi' ? 'Perbaiki laporan' : 'Buka tugas'} <span>→</span>
          </a>
        </article>
      {/each}
    </div>
  {:else}
    <section class="panel">
      <div class="empty-state">
        <span>✓</span>
        <strong>Semua tugas tertangani</strong>
        <p>Tugas baru yang ditugaskan manager akan muncul di sini.</p>
      </div>
    </section>
  {/if}
{/if}

