<script>
  import { onMount } from 'svelte';
  import { fetchAssets, deleteAsset } from './lib/api.js';

  let assets = [];
  let loading = true;
  let error = '';

  async function loadData() {
    loading = true;
    try {
      assets = await fetchAssets();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function handleDelete(id) {
    if (confirm('Apakah Anda yakin ingin menghapus mesin ini?')) {
      await deleteAsset(id);
      loadData();
    }
  }

  onMount(() => {
    loadData();
  });
</script>

<main class="container">
  <header>
    <h1>⚙️ Smart Maintenance - Master Asset</h1>
    <button class="btn-primary" on:click={loadData}>Refresh Data</button>
  </header>

  {#if loading}
    <p class="status">Memuat data dari database MySQL...</p>
  {:else if error}
    <p class="status error">Error: {error}</p>
  {:else}
    <table class="data-table">
      <thead>
        <tr>
          <th>Kode Asset</th>
          <th>Nama Mesin</th>
          <th>Merk / Model</th>
          <th>Lokasi</th>
          <th>Kondisi</th>
          <th>Status</th>
          <th>Aksi</th>
        </tr>
      </thead>
      <tbody>
        {#each assets as item}
          <tr>
            <td><strong>{item.kode_asset}</strong></td>
            <td>{item.nama_mesin}</td>
            <td>{item.merk || '-'} {item.tipe_model ? `(${item.tipe_model})` : ''}</td>
            <td>{item.lokasi || '-'}</td>
            <td>
              <span class="badge {item.kondisi === 'Baik' ? 'green' : 'red'}">
                {item.kondisi}
              </span>
            </td>
            <td>
              <span class="badge {item.status === 'Operasional' ? 'blue' : 'yellow'}">
                {item.status}
              </span>
            </td>
            <td>
              <button class="btn-danger" on:click={() => handleDelete(item.id_asset)}>Hapus</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</main>

<style>
  .container { max-width: 1100px; margin: 2rem auto; font-family: system-ui, sans-serif; }
  header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
  .data-table { width: 100%; border-collapse: collapse; text-align: left; }
  .data-table th, .data-table td { padding: 12px; border-bottom: 1px solid #e2e8f0; }
  .data-table th { background-color: #f8fafc; color: #475569; }
  .badge { padding: 4px 8px; border-radius: 4px; font-size: 0.85rem; font-weight: 600; }
  .green { background: #dcfce7; color: #166534; }
  .red { background: #fee2e2; color: #991b1b; }
  .blue { background: #dbeafe; color: #1e40af; }
  .yellow { background: #fef9c3; color: #854d0e; }
  .btn-primary { background: #2563eb; color: white; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; }
  .btn-danger { background: #ef4444; color: white; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; }
</style>