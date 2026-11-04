<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'

  let users = $state([])
  let loading = $state(true)
  let flashMsg = $state({ type: '', text: '' })

  let newUser = $state({
    full_name: '',
    username: '',
    password: '',
    role: 'Teknisi',
  })
  let creating = $state(false)

  async function loadUsers() {
    loading = true
    try {
      users = await api.users.list()
    } catch (err) {
      flashMsg = { type: 'error', text: err.message || 'Gagal memuat pengguna' }
    } finally {
      loading = false
    }
  }

  async function handleCreateUser(e) {
    e.preventDefault()
    if (!newUser.full_name || !newUser.username || !newUser.password) {
      alert('Nama lengkap, username, dan kata sandi wajib diisi.')
      return
    }
    if (newUser.password.length < 8) {
      alert('Kata sandi minimal 8 karakter.')
      return
    }

    creating = true
    try {
      const res = await api.users.create(newUser)
      flashMsg = { type: 'success', text: res.message || 'Akun berhasil dibuat.' }
      newUser = { full_name: '', username: '', password: '', role: 'Teknisi' }
      await loadUsers()
    } catch (err) {
      flashMsg = { type: 'error', text: err.message || 'Gagal membuat akun' }
    } finally {
      creating = false
    }
  }

  onMount(() => {
    loadUsers()
  })
</script>

<div class="page-heading">
  <div>
    <span class="eyebrow">AKSES SISTEM</span>
    <h1>Pengguna</h1>
    <p>Kelola akun manager dan teknisi untuk akses workspace.</p>
  </div>
</div>

{#if flashMsg.text}
  <div class="flash-stack">
    <div class="flash {flashMsg.type}">{flashMsg.text}</div>
  </div>
{/if}

<div class="split-layout">
  <!-- Left: Users List -->
  <section class="panel">
    <div class="panel-heading">
      <div>
        <h2>Daftar pengguna</h2>
        <p>{users.length} akun terdaftar</p>
      </div>
    </div>

    {#if loading}
      <div class="empty-state">
        <span>♙</span>
        <strong>Memuat data pengguna...</strong>
      </div>
    {:else}
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Nama</th>
              <th>Username</th>
              <th>Peran</th>
              <th>Tugas aktif</th>
            </tr>
          </thead>
          <tbody>
            {#each users as user}
              <tr>
                <td><strong>{user.full_name}</strong></td>
                <td><span class="code-chip">{user.username}</span></td>
                <td>
                  <span
                    class="status-pill {user.role === 'Manager' ? 'status-info' : 'status-muted'}"
                  >
                    {user.role}
                  </span>
                </td>
                <td>{user.assigned_count || 0}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </section>

  <!-- Right: Create User Form -->
  <section class="panel">
    <div class="panel-heading">
      <div>
        <h2>Tambah akun</h2>
        <p>Buat akun untuk anggota tim.</p>
      </div>
    </div>

    <form onsubmit={handleCreateUser} class="stack-form compact-form">
      <label>
        Nama lengkap
        <input
          bind:value={newUser.full_name}
          required
          maxlength="120"
          placeholder="Nama pengguna"
        />
      </label>
      <label>
        Username
        <input
          bind:value={newUser.username}
          required
          maxlength="60"
          autocomplete="off"
          placeholder="username"
        />
      </label>
      <label>
        Kata sandi
        <input
          type="password"
          bind:value={newUser.password}
          required
          minlength="8"
          autocomplete="new-password"
          placeholder="Minimal 8 karakter"
        />
      </label>
      <label>
        Peran
        <select bind:value={newUser.role} required>
          <option value="Teknisi">Teknisi</option>
          <option value="Manager">Manager</option>
        </select>
      </label>
      <button
        class="button primary full-button"
        type="submit"
        disabled={creating}
      >
        {creating ? 'Membuat akun...' : 'Buat akun'} <span>→</span>
      </button>
    </form>
  </section>
</div>

