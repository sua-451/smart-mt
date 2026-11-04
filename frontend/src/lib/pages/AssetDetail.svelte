<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'
  import { statusClass, formatDate } from '../utils.js'

  let { assetId, currentUser } = $props()

  let data = $state(null)
  let history = $state([])
  let loading = $state(true)
  let errorMsg = $state('')

  async function loadDetail() {
    loading = true
    errorMsg = ''
    try {
      const res = await api.assets.get(assetId)
      data = res
      if (res && res.history) {
        history = res.history
      } else {
        const allOrders = await api.workOrders.list()
        history = (allOrders || []).filter(
          (o) => String(o.id_asset || o.asset_id) === String(assetId)
        )
      }
    } catch (err) {
      errorMsg = err.message || 'Gagal memuat detail aset'
    } finally {
      loading = false
    }
  }

  onMount(() => {
    loadDetail()
  })

  let asset = $derived(data?.asset || {})
  let photos = $derived(data?.photos || [])
  let pmHistory = $derived(data?.pm_history || [])
  let failureHistory = $derived(data?.failure_history || [])
  let conditionHistory = $derived(data?.condition_history || [])
  let downtimeHistory = $derived(data?.downtime_history || [])
  let kpiHistory = $derived(data?.kpi_history || [])
</script>

{#if loading}
  <section class="panel">
    <div class="empty-state">
      <span>⬡</span>
      <strong>Memuat detail aset...</strong>
    </div>
  </section>
{:else if errorMsg}
  <div class="flash-stack">
    <div class="flash error">{errorMsg}</div>
  </div>
  <a class="button secondary" href="#/assets">← Kembali ke daftar aset</a>
{:else if data}
  <div class="page-heading asset-detail-heading">
    <div>
      <span class="eyebrow">DETAIL MASTER_ASSET · #{asset.id_asset}</span>
      <h1>{asset.nama_mesin}</h1>
      <p>
        {asset.kode_asset}{asset.serial_number ? ` · SN ${asset.serial_number}` : ''}
      </p>
    </div>
    <div class="asset-detail-actions">
      <a class="button secondary" href="#/assets">← Kembali ke daftar aset</a>
      {#if currentUser?.role === 'Manager'}
        <a class="button primary" href="#/assets/{asset.id_asset}/edit">Edit aset</a>
      {/if}
    </div>
  </div>

  <div class="asset-detail-grid">
    <!-- Left: Asset Information Panel -->
    <section class="panel asset-information">
      <div class="panel-heading">
        <div>
          <h2>Informasi mesin</h2>
          <p>Informasi dan foto aset terdaftar.</p>
        </div>
        <span class="status-pill {statusClass(asset.status)}">{asset.status}</span>
      </div>
      <dl class="asset-info-list">
        <div>
          <dt>Kode aset</dt>
          <dd><span class="code-chip">{asset.kode_asset}</span></dd>
        </div>
        <div>
          <dt>Serial number</dt>
          <dd>{asset.serial_number || '—'}</dd>
        </div>
        <div>
          <dt>Merek</dt>
          <dd>{asset.merk || '—'}</dd>
        </div>
        <div>
          <dt>Tipe / model</dt>
          <dd>{asset.tipe_model || '—'}</dd>
        </div>
        <div>
          <dt>Tahun perolehan</dt>
          <dd>{asset.tahun_perolehan || '—'}</dd>
        </div>
        <div>
          <dt>Kondisi</dt>
          <dd>{asset.kondisi || '—'}</dd>
        </div>
        <div>
          <dt>Lokasi</dt>
          <dd>{asset.lokasi || '—'}</dd>
        </div>
        <div>
          <dt>Penanggung jawab</dt>
          <dd>{asset.penanggung_jawab || '—'}</dd>
        </div>
        <div>
          <dt>Spesifikasi</dt>
          <dd>{asset.kapasitas_spesifikasi || '—'}</dd>
        </div>
        <div>
          <dt>Sumber data</dt>
          <dd>{asset.sumber_data || '—'}</dd>
        </div>
        <div>
          <dt>Tanggal pendataan</dt>
          <dd>{asset.tanggal_pendataan || '—'}</dd>
        </div>
        <div>
          <dt>Petugas</dt>
          <dd>{asset.petugas || '—'}</dd>
        </div>
        <div>
          <dt>Total work order</dt>
          <dd>{history.length} riwayat</dd>
        </div>
      </dl>

      <div class="asset-detail-gallery-heading">
        <strong>Foto mesin</strong>
        <span>{photos.length} foto</span>
      </div>

      {#if photos && photos.length > 0}
        <div class="asset-detail-gallery">
          {#each photos as photo}
            <a href={api.uploadUrl(photo.filename)} target="_blank" rel="noopener">
              <img
                src={api.uploadUrl(photo.filename)}
                alt="Foto {asset.nama_mesin}"
                loading="lazy"
              />
            </a>
          {/each}
        </div>
      {:else}
        <div class="asset-gallery-empty">Belum ada foto untuk mesin ini.</div>
      {/if}
    </section>

    <!-- Right: Work Order Maintenance History -->
    <section class="panel asset-history-panel">
      <div class="panel-heading">
        <div>
          <h2>Riwayat maintenance</h2>
          <p>Jadwal dan laporan maintenance mesin ini, dari yang terbaru.</p>
        </div>
        <span class="history-count">{history.length} tiket</span>
      </div>

      {#if history && history.length > 0}
        <div class="asset-history-list">
          {#each history as entry}
            <article class="history-card">
              <div class="history-card-header">
                <div>
                  <span class="history-date">
                    {entry.work_date || entry.scheduled_date}
                  </span>
                  <h3>
                    {entry.maintenance_type}
                    <span>· {entry.ticket_id || entry.wo_number}</span>
                  </h3>
                </div>
                <span class="status-pill {statusClass(entry.status)}">
                  {entry.status}
                </span>
              </div>

              <div class="history-meta">
                <span>Jadwal {entry.scheduled_date}</span>
                <span>Tingkat {entry.severity}</span>
                <span>Teknisi {entry.technician_name || 'Belum ditugaskan'}</span>
              </div>

              {#if entry.work_date}
                <div class="history-conditions">
                  <div>
                    <strong>Kondisi sebelum</strong>
                    <p>{entry.before_condition}</p>
                    {#if entry.before_photo}
                      <a
                        class="evidence-link"
                        href={api.uploadUrl(entry.before_photo)}
                        target="_blank"
                        rel="noopener"
                      >
                        Lihat foto ↗
                      </a>
                    {/if}
                  </div>
                  <div>
                    <strong>Kondisi sesudah</strong>
                    <p>{entry.after_condition}</p>
                    {#if entry.after_photo}
                      <a
                        class="evidence-link"
                        href={api.uploadUrl(entry.after_photo)}
                        target="_blank"
                        rel="noopener"
                      >
                        Lihat foto ↗
                      </a>
                    {/if}
                  </div>
                </div>

                {#if entry.report_details && entry.report_details.length > 0}
                  {#each entry.report_details as detail, dIdx}
                    <div class="history-repair-detail">
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

                {#if entry.manager_note}
                  <div class="history-manager-note">
                    <strong>Catatan manager</strong>
                    <p>{entry.manager_note}</p>
                  </div>
                {/if}
              {:else}
                <p class="history-pending">Laporan pengerjaan belum tersedia.</p>
              {/if}
            </article>
          {/each}
        </div>
      {:else}
        <div class="empty-state">
          <span>◷</span>
          <strong>Belum ada riwayat maintenance</strong>
          <p>Riwayat work order untuk mesin ini akan ditampilkan di sini.</p>
        </div>
      {/if}
    </section>
  </div>

  <!-- PM Schedule table -->
  {#if pmHistory && pmHistory.length > 0}
    <section class="panel asset-native-history" style="margin-top: 16px;">
      <div class="panel-heading">
        <div>
          <h2>Jadwal preventive maintenance · pm_schedule</h2>
          <p>Jadwal PM dari database.</p>
        </div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>ID PM</th>
              <th>Item PM</th>
              <th>Aktivitas</th>
              <th>Interval</th>
              <th>Terjadwal</th>
              <th>Realisasi</th>
              <th>Petugas</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {#each pmHistory as pm}
              <tr>
                <td><span class="code-chip">{pm.pm_id}</span></td>
                <td>{pm.item_pm}</td>
                <td>{pm.aktivitas?.String || pm.aktivitas || '—'}</td>
                <td>{pm.interval?.String || pm.interval || '—'}</td>
                <td>{pm.tanggal_terjadwal}</td>
                <td>{pm.tanggal_realisasi?.String || pm.tanggal_realisasi || '—'}</td>
                <td>{pm.petugas?.String || pm.petugas || '—'}</td>
                <td>
                  <span class="status-pill {statusClass(pm.status)}">{pm.status}</span>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </section>
  {/if}

  <!-- Failure History -->
  {#if failureHistory && failureHistory.length > 0}
    <section class="panel asset-native-history" style="margin-top: 16px;">
      <div class="panel-heading">
        <div>
          <h2>Histori kerusakan · histori_kerusakan</h2>
          <p>Data kerusakan yang tidak terkait dengan work order.</p>
        </div>
      </div>
      <div class="asset-history-list">
        {#each failureHistory as failure}
          <article class="history-card">
            <div class="history-card-header">
              <div>
                <span class="history-date">{failure.tanggal_jam_mulai}</span>
                <h3>{failure.komponen?.String || failure.komponen || failure.failure_id}</h3>
              </div>
            </div>
            <div class="history-conditions">
              <div>
                <strong>Gejala</strong>
                <p>{failure.gejala?.String || failure.gejala || '—'}</p>
              </div>
              <div>
                <strong>Penyebab</strong>
                <p>{failure.penyebab?.String || failure.penyebab || '—'}</p>
              </div>
              <div>
                <strong>Tindakan perbaikan</strong>
                <p>{failure.tindakan_perbaikan?.String || failure.tindakan_perbaikan || '—'}</p>
              </div>
              <div>
                <strong>Teknisi</strong>
                <p>{failure.teknisi?.String || failure.teknisi || '—'}</p>
              </div>
              <div>
                <strong>Repair time</strong>
                <p>{failure.repair_time_jam?.Float64 ?? failure.repair_time_jam ?? '—'} jam</p>
              </div>
              <div>
                <strong>Downtime</strong>
                <p>{failure.downtime_jam?.Float64 ?? failure.downtime_jam ?? '—'} jam</p>
              </div>
            </div>
            {#if failure.bukti_link?.String || failure.bukti_link}
              <p class="history-pending">
                Bukti: {failure.bukti_link?.String || failure.bukti_link}
              </p>
            {/if}
          </article>
        {/each}
      </div>
    </section>
  {/if}

  <!-- Condition History -->
  {#if conditionHistory && conditionHistory.length > 0}
    <section class="panel asset-native-history" style="margin-top: 16px;">
      <div class="panel-heading">
        <div>
          <h2>Kondisi awal · kondisi_awal</h2>
          <p>Checklist dan kondisi aktual aset.</p>
        </div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Tanggal</th>
              <th>Komponen</th>
              <th>Standar</th>
              <th>Kondisi aktual</th>
              <th>OK/NG</th>
              <th>Temuan</th>
              <th>Tindakan</th>
              <th>Petugas</th>
            </tr>
          </thead>
          <tbody>
            {#each conditionHistory as condition}
              <tr>
                <td>{condition.tanggal}</td>
                <td>{condition.komponen_item}</td>
                <td>{condition.standar_kondisi?.String || condition.standar_kondisi || '—'}</td>
                <td>{condition.kondisi_aktual?.String || condition.kondisi_aktual || '—'}</td>
                <td>{condition.ok_ng?.String || condition.ok_ng || '—'}</td>
                <td>{condition.temuan?.String || condition.temuan || '—'}</td>
                <td>{condition.tindakan?.String || condition.tindakan || '—'}</td>
                <td>{condition.petugas?.String || condition.petugas || '—'}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </section>
  {/if}

  <!-- Downtime History -->
  {#if downtimeHistory && downtimeHistory.length > 0}
    <section class="panel asset-native-history" style="margin-top: 16px;">
      <div class="panel-heading">
        <div>
          <h2>Downtime</h2>
          <p>Catatan waktu mesin tidak beroperasi.</p>
        </div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Tanggal</th>
              <th>Jenis</th>
              <th>Mulai</th>
              <th>Selesai</th>
              <th>Durasi</th>
              <th>Penyebab</th>
              <th>Work order</th>
            </tr>
          </thead>
          <tbody>
            {#each downtimeHistory as dt}
              <tr>
                <td>{dt.downtime_id}</td>
                <td>{dt.tanggal}</td>
                <td>{dt.jenis?.String || dt.jenis || '—'}</td>
                <td>{dt.mulai_downtime?.String || dt.mulai_downtime || '—'}</td>
                <td>{dt.selesai_downtime?.String || dt.selesai_downtime || '—'}</td>
                <td>{dt.durasi_jam?.Float64 ?? dt.durasi_jam ?? '—'} jam</td>
                <td>{dt.penyebab?.String || dt.penyebab || '—'}</td>
                <td>{dt.wo_number?.String || dt.wo_number || '—'}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </section>
  {/if}

  <!-- KPI History -->
  {#if kpiHistory && kpiHistory.length > 0}
    <section class="panel asset-native-history" style="margin-top: 16px;">
      <div class="panel-heading">
        <div>
          <h2>KPI before/after</h2>
          <p>Nilai indikator sesuai data kpi_before_after.</p>
        </div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Periode</th>
              <th>Operating time</th>
              <th>Failure</th>
              <th>Repair time</th>
              <th>Downtime</th>
              <th>MTBF</th>
              <th>MTTR</th>
              <th>Availability</th>
              <th>PM compliance</th>
            </tr>
          </thead>
          <tbody>
            {#each kpiHistory as kpi}
              <tr>
                <td>{kpi.periode}</td>
                <td>{kpi.operating_time_jam ?? '—'}</td>
                <td>{kpi.failure ?? '—'}</td>
                <td>{kpi.repair_time_jam ?? '—'}</td>
                <td>{kpi.downtime_jam ?? '—'}</td>
                <td>{kpi.mtbf_jam != null ? kpi.mtbf_jam : '—'}</td>
                <td>{kpi.mttr_jam != null ? kpi.mttr_jam : '—'}</td>
                <td>{kpi.availability_persen != null ? `${kpi.availability_persen}%` : '—'}</td>
                <td>{kpi.pm_compliance_persen != null ? `${kpi.pm_compliance_persen}%` : '—'}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </section>
  {/if}
{/if}

