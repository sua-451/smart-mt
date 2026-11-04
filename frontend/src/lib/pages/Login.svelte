<script>
  import { api, setAuth } from '../api.js'

  let { onLoginSuccess } = $props()

  let username = $state('')
  let password = $state('')
  let errorMsg = $state('')
  let loading = $state(false)

  async function handleSubmit(e) {
    e.preventDefault()
    errorMsg = ''
    loading = true
    try {
      const res = await api.login(username, password)
      setAuth(res.token, res.user)
      if (onLoginSuccess) {
        onLoginSuccess(res.user)
      } else {
        window.location.hash = '#/dashboard'
      }
    } catch (err) {
      errorMsg = err.message || 'Username atau kata sandi tidak sesuai.'
    } finally {
      loading = false
    }
  }
</script>

{#if errorMsg}
  <div class="login-flashes">
    <div class="flash error">{errorMsg}</div>
  </div>
{/if}

<main class="login-page">
  <section class="login-visual">
    <a class="brand brand-on-dark" href="#/login">
      <span class="brand-mark">SM</span>
      <span>SMART<span class="brand-light">Maintenance</span></span>
    </a>
    <div class="visual-copy">
      <span class="eyebrow">MAINTENANCE WORKSPACE</span>
      <h1>Operasional lancar dimulai dari mesin yang terawat.</h1>
      <p>Jadwalkan, dokumentasikan, dan pantau setiap proses maintenance dari satu tempat.</p>
    </div>
    <div class="visual-footer">Sistem Manajemen Maintenance · PNC</div>
    <div class="visual-orbit orbit-one"></div>
    <div class="visual-orbit orbit-two"></div>
  </section>
  <section class="login-panel">
    <div class="login-form-wrap">
      <div class="mobile-brand brand">
        <span class="brand-mark">SM</span>
        <span>SMART<span class="brand-light">Maintenance</span></span>
      </div>
      <span class="eyebrow">SELAMAT DATANG</span>
      <h2>Masuk ke akun Anda</h2>
      <p class="muted">Gunakan akun yang sudah terdaftar untuk melanjutkan.</p>
      <form onsubmit={handleSubmit} class="stack-form login-form">
        <label>
          Username
          <input
            bind:value={username}
            autocomplete="username"
            required
            placeholder="Masukkan username"
          />
        </label>
        <label>
          Kata sandi
          <input
            type="password"
            bind:value={password}
            autocomplete="current-password"
            required
            placeholder="Masukkan kata sandi"
          />
        </label>
        <button class="button primary full-button" type="submit" disabled={loading}>
          {loading ? 'Memproses...' : 'Masuk ke workspace'} <span>→</span>
        </button>
      </form>
      <div class="demo-note">
        <strong>Akun demo</strong>
        <span>Manager: <code>manager</code> / <code>manager123</code></span>
        <span>Teknisi: <code>teknisi</code> / <code>teknisi123</code></span>
      </div>
      <p class="login-footnote">Ganti kredensial demo sebelum penggunaan produksi.</p>
    </div>
  </section>
</main>

