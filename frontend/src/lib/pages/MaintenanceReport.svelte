<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'
  import { statusClass, today } from '../utils.js'

  let { ticketId, currentUser } = $props()

  let order = $state(null)
  let loading = $state(true)
  let submitting = $state(false)
  let errorMsg = $state('')
  let flashMsg = $state({ type: '', text: '' })

  let workDate = $state(today())
  let beforeCondition = $state('')
  let beforePhotoFile = $state(null)
  let afterCondition = $state('')
  let afterPhotoFile = $state(null)

  let existingBeforePhoto = $state('')
  let existingAfterPhoto = $state('')
  let managerNote = $state('')

  let details = $state([
    {
      damage_description: '',
      damage_photo_file: null,
      existing_damage_photo: '',
      repair_action: '',
      repair_photo_file: null,
      existing_repair_photo: '',
    },
  ])

  async function loadOrder() {
    loading = true
    errorMsg = ''
    try {
      const res = await api.workOrders.get(ticketId)
      order = res
      if (res) {
        if (res.work_date) workDate = res.work_date
        if (res.before_condition) beforeCondition = res.before_condition
        if (res.after_condition) afterCondition = res.after_condition
        if (res.before_photo) existingBeforePhoto = res.before_photo
        if (res.after_photo) existingAfterPhoto = res.after_photo
        if (res.manager_note) managerNote = res.manager_note

        if (res.report_details && res.report_details.length > 0) {
          details = res.report_details.map((d) => ({
            damage_description: d.damage_description || '',
            damage_photo_file: null,
            existing_damage_photo: d.damage_photo?.String || d.damage_photo || '',
            repair_action: d.repair_action || '',
            repair_photo_file: null,
            existing_repair_photo: d.repair_photo?.String || d.repair_photo || '',
          }))
        }
      }
    } catch (err) {
      errorMsg = err.message || 'Gagal memuat data laporan'
    } finally {
      loading = false
    }
  }

  function addDetail() {
    details = [
      ...details,
      {
        damage_description: '',
        damage_photo_file: null,
        existing_damage_photo: '',
        repair_action: '',
        repair_photo_file: null,
        existing_repair_photo: '',
      },
    ]
  }

  function removeDetail(index) {
    if (details.length <= 1) return
    details = details.filter((_, i) => i !== index)
  }

  async function handleSubmit(e) {
    e.preventDefault()
    if (!beforeCondition.trim() || !afterCondition.trim()) {
      alert('Kondisi sebelum dan sesudah wajib diisi.')
      return
    }

    submitting = true
    const fd = new FormData()
    fd.append('work_date', workDate)
    fd.append('before_condition', beforeCondition)
    fd.append('after_condition', afterCondition)

    if (beforePhotoFile) {
      fd.append('before_photo', beforePhotoFile)
    }
    if (afterPhotoFile) {
      fd.append('after_photo', afterPhotoFile)
    }

    details.forEach((d, i) => {
      fd.append('damage_description[]', d.damage_description)
      fd.append('repair_action[]', d.repair_action)
      if (d.damage_photo_file) {
        fd.append(`damage_photo_${i}`, d.damage_photo_file)
      } else if (d.existing_damage_photo) {
        fd.append(`existing_damage_photo_${i}`, d.existing_damage_photo)
      }
      if (d.repair_photo_file) {
        fd.append(`repair_photo_${i}`, d.repair_photo_file)
      } else if (d.existing_repair_photo) {
        fd.append(`existing_repair_photo_${i}`, d.existing_repair_photo)
      }
    })

    try {
      const res = await api.workOrders.report(ticketId, fd)
      flashMsg = { type: 'success', text: res.message || 'Laporan berhasil dikirim.' }
      window.location.hash = '#/dashboard'
    } catch (err) {
      flashMsg = { type: 'error', text: err.message || 'Gagal mengirim laporan' }
    } finally {
      submitting = false
    }
  }

  onMount(() => {
    loadOrder()
  })
</script>

{#if loading}
  <section class="panel">
    <div class="empty-state">
      <span>◷</span>
      <strong>Memuat formulir laporan...</strong>
    </div>
  </section>
{:else if errorMsg}
  <div class="flash-stack">
    <div class="flash error">{errorMsg}</div>
  </div>
  <a class="button secondary" href="#/dashboard">← Kembali</a>
{:else if order}
  <div class="page-heading">
    <div>
      <span class="eyebrow">
        TIKET #{order.ticket_id || order.wo_number} · {(order.maintenance_type || '').toUpperCase()}
      </span>
      <h1>{order.machine_name}</h1>
      <p>
        {order.asset_id} · {order.category} · {order.location} · Jadwal {order.scheduled_date}
      </p>
    </div>
    <span class="status-pill {statusClass(order.status)}">{order.status}</span>
  </div>

  {#if flashMsg.text}
    <div class="flash-stack">
      <div class="flash {flashMsg.type}">{flashMsg.text}</div>
    </div>
  {/if}

  {#if managerNote}
    <div class="revision-banner">
      <strong>Catatan revisi dari manager</strong>
      <p>{managerNote}</p>
    </div>
  {/if}

  <form onsubmit={handleSubmit} class="maintenance-form" id="maintenance-form">
    <!-- SECTION 01: Kondisi Mesin -->
    <section class="panel report-section">
      <div class="section-number">01</div>
      <div class="report-section-content">
        <div class="panel-heading">
          <div>
            <h2>Kondisi mesin</h2>
            <p>Catat kondisi keseluruhan sebelum dan sesudah pekerjaan.</p>
          </div>
        </div>
        <div class="form-grid">
          <label>
            Tanggal pengerjaan
            <input type="date" bind:value={workDate} required />
          </label>
          <div class="form-spacer"></div>

          <label>
            Kondisi sebelum <span class="required">*</span>
            <textarea
              bind:value={beforeCondition}
              rows="4"
              required
              placeholder="Jelaskan kondisi mesin saat ditemukan"
            ></textarea>
          </label>
          <label>
            Foto kondisi sebelum
            <input
              type="file"
              accept=".jpg,.jpeg,.png,.webp,image/jpeg,image/png,image/webp"
              onchange={(e) => (beforePhotoFile = e.target.files[0])}
            />
            {#if existingBeforePhoto}
              <a
                class="evidence-link"
                href={api.uploadUrl(existingBeforePhoto)}
                target="_blank"
                rel="noopener"
              >
                Lihat foto sebelumnya ↗
              </a>
            {/if}
          </label>

          <label>
            Kondisi sesudah <span class="required">*</span>
            <textarea
              bind:value={afterCondition}
              rows="4"
              required
              placeholder="Jelaskan kondisi mesin setelah maintenance"
            ></textarea>
          </label>
          <label>
            Foto kondisi sesudah
            <input
              type="file"
              accept=".jpg,.jpeg,.png,.webp,image/jpeg,image/png,image/webp"
              onchange={(e) => (afterPhotoFile = e.target.files[0])}
            />
            {#if existingAfterPhoto}
              <a
                class="evidence-link"
                href={api.uploadUrl(existingAfterPhoto)}
                target="_blank"
                rel="noopener"
              >
                Lihat foto sebelumnya ↗
              </a>
            {/if}
          </label>
        </div>
      </div>
    </section>

    <!-- SECTION 02: Detail Kerusakan & Tindakan -->
    <section class="panel report-section">
      <div class="section-number">02</div>
      <div class="report-section-content">
        <div class="panel-heading">
          <div>
            <h2>Detail kerusakan &amp; tindakan</h2>
            <p>Tambahkan kartu untuk setiap titik kerusakan yang ditemukan.</p>
          </div>
          <button
            class="button secondary"
            type="button"
            onclick={addDetail}
          >
            ＋ Tambah kartu
          </button>
        </div>

        <div id="detail-cards" class="detail-card-list">
          {#each details as detail, idx}
            <article class="detail-card">
              <div class="detail-card-heading">
                <strong>Detail {idx + 1}</strong>
                <button
                  class="icon-button remove-detail"
                  type="button"
                  aria-label="Hapus kartu"
                  disabled={details.length <= 1}
                  onclick={() => removeDetail(idx)}
                >
                  ×
                </button>
              </div>
              <div class="form-grid">
                <label>
                  Detail kerusakan <span class="required">*</span>
                  <textarea
                    bind:value={detail.damage_description}
                    rows="3"
                    required
                    placeholder="Jelaskan titik kerusakan"
                  ></textarea>
                </label>
                <label>
                  Foto kerusakan
                  <input
                    type="file"
                    accept=".jpg,.jpeg,.png,.webp,image/jpeg,image/png,image/webp"
                    onchange={(e) => (detail.damage_photo_file = e.target.files[0])}
                  />
                  {#if detail.existing_damage_photo}
                    <a
                      class="evidence-link"
                      href={api.uploadUrl(detail.existing_damage_photo)}
                      target="_blank"
                      rel="noopener"
                    >
                      Lihat foto tersimpan ↗
                    </a>
                  {/if}
                </label>

                <label>
                  Tindakan perbaikan <span class="required">*</span>
                  <textarea
                    bind:value={detail.repair_action}
                    rows="3"
                    required
                    placeholder="Jelaskan tindakan yang dilakukan"
                  ></textarea>
                </label>
                <label>
                  Foto hasil perbaikan
                  <input
                    type="file"
                    accept=".jpg,.jpeg,.png,.webp,image/jpeg,image/png,image/webp"
                    onchange={(e) => (detail.repair_photo_file = e.target.files[0])}
                  />
                  {#if detail.existing_repair_photo}
                    <a
                      class="evidence-link"
                      href={api.uploadUrl(detail.existing_repair_photo)}
                      target="_blank"
                      rel="noopener"
                    >
                      Lihat foto tersimpan ↗
                    </a>
                  {/if}
                </label>
              </div>
            </article>
          {/each}
        </div>
      </div>
    </section>

    <div class="form-actions sticky-actions">
      <a class="button secondary" href="#/dashboard">Kembali</a>
      <button class="button primary" type="submit" disabled={submitting}>
        {submitting ? 'Mengirim...' : 'Kirim laporan'} <span>→</span>
      </button>
    </div>
  </form>
{/if}

