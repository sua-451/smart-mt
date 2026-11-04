<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'

  let { assetId, currentUser } = $props()

  let editing = $derived(!!assetId)

  let formData = $state({
    kode_asset: '',
    nama_mesin: '',
    merk: '',
    tipe_model: '',
    serial_number: '',
    tahun_perolehan: '',
    lokasi: '',
    penanggung_jawab: '',
    status: 'Operasional',
    kondisi: 'Baik',
    tanggal_pendataan: '',
    petugas: '',
    kapasitas_spesifikasi: '',
    sumber_data: '',
  })

  let existingPhotos = $state([])
  let removePhotoIds = $state([])
  let photoFiles = $state([])
  let loading = $state(false)
  let submitting = $state(false)
  let errorMsg = $state('')

  async function loadAssetData() {
    if (!assetId) return
    loading = true
    try {
      const res = await api.assets.get(assetId)
      if (res && res.asset) {
        const a = res.asset
        formData = {
          kode_asset: a.kode_asset || '',
          nama_mesin: a.nama_mesin || '',
          merk: a.merk || '',
          tipe_model: a.tipe_model || '',
          serial_number: a.serial_number || '',
          tahun_perolehan: a.tahun_perolehan ? String(a.tahun_perolehan) : '',
          lokasi: a.lokasi || '',
          penanggung_jawab: a.penanggung_jawab || '',
          status: a.status || 'Operasional',
          kondisi: a.kondisi || 'Baik',
          tanggal_pendataan: a.tanggal_pendataan || '',
          petugas: a.petugas || '',
          kapasitas_spesifikasi: a.kapasitas_spesifikasi || '',
          sumber_data: a.sumber_data || '',
        }
        existingPhotos = res.photos || []
      }
    } catch (err) {
      errorMsg = err.message || 'Gagal memuat data aset'
    } finally {
      loading = false
    }
  }

  function handleFileChange(e) {
    photoFiles = Array.from(e.target.files || [])
  }

  function toggleRemovePhoto(photoId) {
    if (removePhotoIds.includes(photoId)) {
      removePhotoIds = removePhotoIds.filter((id) => id !== photoId)
    } else {
      removePhotoIds = [...removePhotoIds, photoId]
    }
  }

  async function handleSubmit(e) {
    e.preventDefault()
    errorMsg = ''
    submitting = true

    const fd = new FormData()
    for (const [key, value] of Object.entries(formData)) {
      if (value !== null && value !== undefined) {
        fd.append(key, value)
      }
    }

    for (const file of photoFiles) {
      fd.append('photos', file)
    }

    if (removePhotoIds.length > 0) {
      fd.append('remove_photo_ids', removePhotoIds.join(','))
    }

    try {
      if (editing) {
        await api.assets.update(assetId, fd)
      } else {
        await api.assets.create(fd)
      }
      window.location.hash = '#/assets'
    } catch (err) {
      errorMsg = err.message || 'Gagal menyimpan aset'
    } finally {
      submitting = false
    }
  }

  onMount(() => {
    if (!editing && currentUser?.full_name) {
      formData.petugas = currentUser.full_name
    }
    loadAssetData()
  })
</script>

<div class="page-heading">
  <div>
    <span class="eyebrow">MASTER_ASSET</span>
    <h1>{editing ? 'Edit aset' : 'Tambah aset baru'}</h1>
    <p>Informasi mengikuti kolom master_asset dalam skema database Maintenance PNC.</p>
  </div>
  <a class="button secondary" href="#/assets">← Kembali</a>
</div>

{#if errorMsg}
  <div class="flash-stack">
    <div class="flash error">{errorMsg}</div>
  </div>
{/if}

{#if loading}
  <section class="panel">
    <div class="empty-state">
      <span>⬡</span>
      <strong>Memuat form aset...</strong>
    </div>
  </section>
{:else}
  <section class="panel form-panel">
    <form onsubmit={handleSubmit} class="stack-form">
      <div class="form-grid">
        <label>
          Kode aset <span class="required">*</span>
          <input
            bind:value={formData.kode_asset}
            required
            maxlength="30"
            placeholder="Kode dari inventaris"
          />
        </label>
        <label>
          Nama mesin <span class="required">*</span>
          <input
            bind:value={formData.nama_mesin}
            required
            maxlength="100"
            placeholder="Nama mesin"
          />
        </label>
        <label>
          Merek
          <input
            bind:value={formData.merk}
            maxlength="100"
            placeholder="Merek mesin"
          />
        </label>
        <label>
          Tipe / model
          <input
            bind:value={formData.tipe_model}
            maxlength="100"
            placeholder="Tipe atau model"
          />
        </label>
        <label>
          Nomor seri
          <input
            bind:value={formData.serial_number}
            maxlength="50"
            placeholder="Nomor seri fisik mesin"
          />
          <small>Nomor seri bersifat unik jika diisi.</small>
        </label>
        <label>
          Tahun perolehan
          <input
            type="number"
            bind:value={formData.tahun_perolehan}
            min="1900"
            max="2030"
            placeholder="Contoh: 2022"
          />
        </label>
        <label>
          Lokasi
          <input
            bind:value={formData.lokasi}
            maxlength="100"
            placeholder="Lokasi mesin"
          />
        </label>
        <label>
          Penanggung jawab
          <input
            bind:value={formData.penanggung_jawab}
            maxlength="100"
            placeholder="Nama penanggung jawab"
          />
        </label>
        <label>
          Status
          <select bind:value={formData.status} required>
            <option value="Operasional">Operasional</option>
            <option value="Tidak Operasional">Tidak Operasional</option>
            <option value="Perbaikan">Perbaikan</option>
            <option value="Afkir">Afkir</option>
          </select>
        </label>
        <label>
          Kondisi
          <select bind:value={formData.kondisi} required>
            <option value="Baik">Baik</option>
            <option value="Rusak Ringan">Rusak Ringan</option>
            <option value="Rusak Berat">Rusak Berat</option>
          </select>
        </label>
        <label>
          Tanggal pendataan
          <input type="date" bind:value={formData.tanggal_pendataan} />
        </label>
        <label>
          Petugas pendataan
          <input
            bind:value={formData.petugas}
            maxlength="100"
            placeholder="Nama petugas"
          />
        </label>
      </div>

      <label>
        Kapasitas / spesifikasi
        <textarea
          bind:value={formData.kapasitas_spesifikasi}
          rows="3"
          placeholder="Kapasitas dan spesifikasi mesin"
        ></textarea>
      </label>

      <label>
        Sumber data
        <input
          bind:value={formData.sumber_data}
          maxlength="50"
          placeholder="Sumber data aset"
        />
      </label>

      {#if editing && existingPhotos && existingPhotos.length > 0}
        <fieldset class="asset-gallery-fieldset">
          <legend>Foto mesin saat ini</legend>
          <p class="field-description">
            Pilih foto yang ingin dihapus. Foto lainnya akan tetap tersimpan.
          </p>
          <div class="asset-gallery">
            {#each existingPhotos as photo}
              <label class="asset-gallery-item">
                <img
                  src={api.uploadUrl(photo.filename)}
                  alt="Foto mesin"
                  loading="lazy"
                />
                <span>
                  <input
                    type="checkbox"
                    checked={removePhotoIds.includes(photo.photo_id)}
                    onchange={() => toggleRemovePhoto(photo.photo_id)}
                  />
                  Hapus foto
                </span>
              </label>
            {/each}
          </div>
        </fieldset>
      {/if}

      <label class="photo-upload-label">
        {editing ? 'Tambah foto mesin' : 'Foto mesin'}
        <input
          type="file"
          accept=".jpg,.jpeg,.png,.webp,image/jpeg,image/png,image/webp"
          multiple
          onchange={handleFileChange}
        />
        <small>
          Tambahkan satu atau beberapa foto. Format JPG, PNG, atau WebP; total unggahan maksimal 12 MB.
        </small>
      </label>

      <div class="form-actions">
        <a class="button secondary" href="#/assets">Batal</a>
        <button class="button primary" type="submit" disabled={submitting}>
          {submitting
            ? 'Menyimpan...'
            : editing
            ? 'Simpan perubahan'
            : 'Simpan aset'}
        </button>
      </div>
    </form>
  </section>
{/if}
