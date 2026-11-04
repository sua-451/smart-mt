<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'
  import { statusClass, today } from '../utils.js'

  let { currentUser } = $props()

  let orders = $state([])
  let assets = $state([])
  let technicians = $state([])
  let loading = $state(true)
  let flashMsg = $state({ type: '', text: '' })

  let newOrder = $state({
    asset_id: '',
    scheduled_date: today(),
    jenis: 'Korektif',
    prioritas: 'Sedang',
    keluhan_gejala: '',
    item_pm: '',
    aktivitas: '',
    interval: '',
    technician_id: '',
  })
  let creatingOrder = $state(false)

  // Track inline assign selections: { [ticket_id]: technician_id }
  let assignSelections = $state({})
  // Track review notes: { [ticket_id]: note }
  let reviewNotes = $state({})

  async function loadData() {
    loading = true
    try {
      const [orderList, assetList, techList] = await Promise.all([
        api.workOrders.list(),
        api.assets.list(),
        api.technicians(),
      ])
      orders = orderList || []
      // Filter active assets for creation
      assets = (assetList || []).filter((a) => a.status !== 'Afkir')
      technicians = techList || []

      // Pre-fill assignSelections
      for (const o of orders) {
        if (o.technician_id) {
          assignSelections[o.ticket_id] = o.technician_id
        }
      }
    } catch (err) {
      flashMsg = { type: 'error', text: err.message || 'Gagal memuat data work order' }
    } finally {
      loading = false
    }
  }

  async function handleAssign(ticketId) {
    const techId = assignSelections[ticketId]
    if (!techId) {
      alert('Pilih teknisi terlebih dahulu')
      return
    }
    try {
      const res = await api.workOrders.assign(ticketId, techId)
      flashMsg = { type: 'success', text: res.message || 'Teknisi berhasil ditugaskan.' }
      await loadData()
    } catch (err) {
      flashMsg = { type: 'error', text: err.message || 'Gagal menugaskan teknisi' }
    }
  }

  async function handleReview(ticketId, action) {
    const note = reviewNotes[ticketId] || ''
    if (action === 'revise' && !note.trim()) {
      alert('Catatan revisi wajib diisi jika meminta revisi.')
      return
    }
    try {
      const res = await api.workOrders.review(ticketId, action, note)
      flashMsg = { type: 'success', text: res.message || 'Laporan berhasil diperbarui.' }
      await loadData()
    } catch (err) {
      flashMsg = { type: 'error', text: err.message || 'Gagal meninjau laporan' }
    }
  }

  async function handleCreateOrder(e) {
    e.preventDefault()
    if (!newOrder.asset_id || !newOrder.scheduled_date || !newOrder.jenis) {
      alert('Aset, tanggal jadwal, dan jenis work order wajib diisi.')
      return
    }
    creatingOrder = true
    try {
      const payload = {
        asset_id: Number(newOrder.asset_id),
        scheduled_date: newOrder.scheduled_date,
        jenis: newOrder.jenis,
        prioritas: newOrder.prioritas,
        keluhan_gejala: newOrder.keluhan_gejala,
        item_pm: newOrder.item_pm,
        aktivitas: newOrder.aktivitas,
        interval: newOrder.interval,
        technician_id: newOrder.technician_id ? Number(newOrder.technician_id) : undefined,
      }
      const res = await api.workOrders.create(payload)
      flashMsg = { type: 'success', text: res.message || 'Work order berhasil dibuat.' }
      newOrder = {
        asset_id: '',
        scheduled_date: today(),
        jenis: 'Korektif',
        prioritas: 'Sedang',
        keluhan_gejala: '',
        item_pm: '',
        aktivitas: '',
        interval: '',
        technician_id: '',
      }
      await loadData()
    } catch (err) {
      flashMsg = { type: 'error', text: err.message || 'Gagal membuat work order' }
    } finally {
      creatingOrder = false
    }
  }

  onMount(() => {
    loadData()
  })
</script>

<div class="page-heading">
  <div>
    <span class="eyebrow">PENJADWALAN &amp; EKSEKUSI</span>
    <h1>Work order</h1>
    <p>Jadwalkan pemeriksaan, tugaskan teknisi, dan tinjau hasil maintenance.</p>
  </div>
</div>

{#if flashMsg.text}
  <div class="flash-stack">
    <div class="flash {flashMsg.type}">{flashMsg.text}</div>
  </div>
{/if}

<div class="split-layout work-order-layout">
  <!-- Left: Work Orders Table -->
  <section class="panel">
    <div class="panel-heading">
      <div>
        <h2>Daftar work order</h2>
        <p>
          {orders.length} tiket dari tabel work_order · jadwal preventif dibuat otomatis setiap bulan
        </p>
      </div>
    </div>

    {#if loading}
      <div class="empty-state">
        <span>◷</span>
        <strong>Memuat work order...</strong>
      </div>
    {:else if orders && orders.length > 0}
      <div class="table-wrap">
        <table class="work-order-table">
          <thead>
            <tr>
              <th>Mesin / Jadwal</th>
              <th>Jenis / Risiko</th>
              <th>Teknisi</th>
              <th>Status &amp; aksi</th>
            </tr>
          </thead>
          <tbody>
            {#each orders as order}
              <tr>
                <td>
                  <strong>{order.machine_name}</strong>
                  <small>{order.asset_id} · {order.scheduled_date}</small>
                </td>
                <td>
                  {order.maintenance_type}
                  <small>{order.severity}</small>
                </td>
                <td>
                  {#if (order.status === 'Menunggu Penugasan' || order.status === 'Ditugaskan') && currentUser?.role === 'Manager'}
                    <div class="inline-assign">
                      <select
                        bind:value={assignSelections[order.ticket_id]}
                        aria-label="Pilih teknisi"
                      >
                        <option value="">Pilih teknisi</option>
                        {#each technicians as tech}
                          <option value={tech.user_id}>{tech.full_name}</option>
                        {/each}
                      </select>
                      <button
                        class="button small secondary"
                        type="button"
                        onclick={() => handleAssign(order.ticket_id)}
                      >
                        Simpan
                      </button>
                    </div>
                  {:else}
                    {order.technician_name || '—'}
                  {/if}
                </td>
                <td>
                  <span class="status-pill {statusClass(order.status)}">
                    {order.status}
                  </span>

                  {#if order.status === 'Menunggu Approval' && order.report_id && currentUser?.role === 'Manager'}
                    <details class="review-details">
                      <summary>Tinjau laporan</summary>
                      <div class="review-card">
                        <div>
                          <strong>Kondisi sebelum</strong>
                          <p>{order.before_condition || '—'}</p>
                        </div>
                        {#if order.before_photo}
                          <a
                            class="evidence-link"
                            href={api.uploadUrl(order.before_photo)}
                            target="_blank"
                            rel="noopener"
                          >
                            Lihat foto kondisi sebelum ↗
                          </a>
                        {/if}
                        <div>
                          <strong>Kondisi sesudah</strong>
                          <p>{order.after_condition || '—'}</p>
                        </div>
                        {#if order.after_photo}
                          <a
                            class="evidence-link"
                            href={api.uploadUrl(order.after_photo)}
                            target="_blank"
                            rel="noopener"
                          >
                            Lihat foto kondisi sesudah ↗
                          </a>
                        {/if}

                        {#if order.report_details && order.report_details.length > 0}
                          {#each order.report_details as detail, dIdx}
                            <div class="review-detail">
                              <strong>Kerusakan {dIdx + 1}</strong>
                              <p>{detail.damage_description}</p>
                              {#if detail.damage_photo}
                                <a
                                  class="evidence-link"
                                  href={api.uploadUrl(detail.damage_photo)}
                                  target="_blank"
                                  rel="noopener"
                                >
                                  Lihat foto kerusakan ↗
                                </a>
                              {/if}
                              <strong>Tindakan perbaikan</strong>
                              <p>{detail.repair_action}</p>
                              {#if detail.repair_photo}
                                <a
                                  class="evidence-link"
                                  href={api.uploadUrl(detail.repair_photo)}
                                  target="_blank"
                                  rel="noopener"
                                >
                                  Lihat foto perbaikan ↗
                                </a>
                              {/if}
                            </div>
                          {/each}
                        {/if}

                        <div class="stack-form" style="margin-top: 10px;">
                          <label>
                            Catatan revisi
                            <textarea
                              bind:value={reviewNotes[order.ticket_id]}
                              rows="2"
                              placeholder="Wajib diisi jika dikembalikan"
                            ></textarea>
                          </label>
                          <div class="row-actions">
                            <button
                              class="button small primary"
                              type="button"
                              onclick={() => handleReview(order.ticket_id, 'approve')}
                            >
                              Setujui
                            </button>
                            <button
                              class="button small danger-outline"
                              type="button"
                              onclick={() => handleReview(order.ticket_id, 'revise')}
                            >
                              Minta revisi
                            </button>
                          </div>
                        </div>
                      </div>
                    </details>
                  {:else if order.status === 'Menunggu Penugasan' && (!technicians || technicians.length === 0)}
                    <small>Tambahkan akun teknisi terlebih dahulu.</small>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <div class="empty-state">
        <span>▤</span>
        <strong>Belum ada work order</strong>
        <p>Jadwalkan tiket manual atau tambahkan aset untuk membuat jadwal rutin.</p>
      </div>
    {/if}
  </section>

  <!-- Right: Create Work Order Form -->
  <section class="panel create-order-panel">
    <div class="panel-heading">
      <div>
        <h2>Buat work order</h2>
        <p>Untuk pekerjaan perbaikan atau jadwal khusus.</p>
      </div>
    </div>

    {#if assets && assets.length > 0}
      <form onsubmit={handleCreateOrder} class="stack-form compact-form">
        <label>
          Aset
          <select bind:value={newOrder.asset_id} required>
            <option value="">Pilih mesin</option>
            {#each assets as asset}
              <option value={asset.id_asset}>
                {asset.kode_asset} · {asset.nama_mesin}
                {asset.serial_number ? ` · SN ${asset.serial_number}` : ''}
              </option>
            {/each}
          </select>
        </label>

        <label>
          Tanggal jadwal
          <input type="date" bind:value={newOrder.scheduled_date} required />
        </label>

        <label>
          Jenis work order
          <select bind:value={newOrder.jenis} required>
            <option value="Korektif">Korektif</option>
            <option value="Preventif">Preventif</option>
            <option value="Predictive">Predictive</option>
            <option value="Lainnya">Lainnya</option>
          </select>
        </label>

        <label>
          Prioritas
          <select bind:value={newOrder.prioritas} required>
            <option value="Rendah">Rendah</option>
            <option value="Sedang">Sedang</option>
            <option value="Tinggi">Tinggi</option>
            <option value="Darurat">Darurat</option>
          </select>
        </label>

        <label>
          Keluhan / gejala
          <textarea
            bind:value={newOrder.keluhan_gejala}
            rows="3"
            placeholder="Deskripsikan keluhan atau pekerjaan yang dijadwalkan"
          ></textarea>
        </label>

        <label>
          Item PM untuk jenis Preventif
          <input
            bind:value={newOrder.item_pm}
            maxlength="150"
            placeholder="Contoh: Pemeriksaan bulanan"
          />
          <small>Kolom item_pm pada pm_schedule.</small>
        </label>

        <label>
          Aktivitas PM
          <textarea
            bind:value={newOrder.aktivitas}
            rows="2"
            placeholder="Aktivitas pemeriksaan atau perawatan"
          ></textarea>
        </label>

        <label>
          Interval PM
          <input
            bind:value={newOrder.interval}
            maxlength="50"
            placeholder="Contoh: 1 bulan"
          />
        </label>

        <label>
          Tugaskan ke teknisi
          <select bind:value={newOrder.technician_id}>
            <option value="">Belum ditugaskan</option>
            {#each technicians as tech}
              <option value={tech.user_id}>{tech.full_name}</option>
            {/each}
          </select>
        </label>

        <button
          class="button primary full-button"
          type="submit"
          disabled={creatingOrder}
        >
          {creatingOrder ? 'Membuat tiket...' : 'Buat tiket'} <span>→</span>
        </button>
      </form>
    {:else}
      <div class="empty-state compact-empty">
        <span>⬡</span>
        <strong>Belum ada aset aktif</strong>
        <p>Tambahkan aset sebelum membuat work order.</p>
        <a class="text-link" href="#/assets/new">Tambah aset →</a>
      </div>
    {/if}
  </section>
</div>

