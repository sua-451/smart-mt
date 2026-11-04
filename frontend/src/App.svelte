<script>
  import { onMount } from 'svelte'
  import { getUser, clearAuth } from './lib/api.js'
  import { formatDate } from './lib/utils.js'

  import Login from './lib/pages/Login.svelte'
  import Dashboard from './lib/pages/Dashboard.svelte'
  import Assets from './lib/pages/Assets.svelte'
  import AssetDetail from './lib/pages/AssetDetail.svelte'
  import AssetForm from './lib/pages/AssetForm.svelte'
  import WorkOrders from './lib/pages/WorkOrders.svelte'
  import MaintenanceReport from './lib/pages/MaintenanceReport.svelte'
  import Users from './lib/pages/Users.svelte'

  let currentUser = $state(getUser())
  let currentHash = $state(window.location.hash || '#/dashboard')

  function handleHashChange() {
    currentHash = window.location.hash || '#/dashboard'
    // If not logged in and not on login, redirect to login
    if (!currentUser && currentHash !== '#/login') {
      window.location.hash = '#/login'
    }
  }

  function handleLoginSuccess(user) {
    currentUser = user
    window.location.hash = '#/dashboard'
  }

  function handleLogout() {
    clearAuth()
    currentUser = null
    window.location.hash = '#/login'
  }

  onMount(() => {
    window.addEventListener('hashchange', handleHashChange)
    if (!currentUser && currentHash !== '#/login') {
      window.location.hash = '#/login'
    }
    return () => {
      window.removeEventListener('hashchange', handleHashChange)
    }
  })

  // Route parser
  let route = $derived.by(() => {
    const raw = currentHash.replace(/^#\/?/, '')
    const parts = raw.split('/').filter(Boolean)

    if (parts.length === 0 || parts[0] === 'dashboard') {
      return { page: 'dashboard' }
    }
    if (parts[0] === 'login') {
      return { page: 'login' }
    }
    if (parts[0] === 'assets') {
      if (parts[1] === 'new') return { page: 'asset_new' }
      if (parts[1] && parts[2] === 'edit') return { page: 'asset_edit', id: parts[1] }
      if (parts[1]) return { page: 'asset_detail', id: parts[1] }
      return { page: 'assets' }
    }
    if (parts[0] === 'work-orders') {
      if (parts[1] && parts[2] === 'report') return { page: 'report', ticketId: parts[1] }
      return { page: 'work_orders' }
    }
    if (parts[0] === 'users') {
      return { page: 'users' }
    }
    return { page: 'dashboard' }
  })

  let breadcrumb = $derived.by(() => {
    switch (route.page) {
      case 'dashboard':
        return 'Dashboard'
      case 'assets':
        return currentUser?.role === 'Manager' ? 'Master Aset' : 'Aset Mesin'
      case 'asset_new':
        return 'Master Aset / Tambah aset'
      case 'asset_detail':
        return 'Aset Mesin / Detail'
      case 'asset_edit':
        return 'Master Aset / Edit aset'
      case 'work_orders':
        return 'Work Order'
      case 'report':
        return 'Tugas saya / Laporan'
      case 'users':
        return 'Pengguna'
      default:
        return 'Dashboard'
    }
  })

  let todayStr = $derived(
    new Date().toLocaleDateString('id-ID', {
      day: '2-digit',
      month: 'long',
      year: 'numeric',
    })
  )
</script>

{#if !currentUser || route.page === 'login'}
  <Login onLoginSuccess={handleLoginSuccess} />
{:else}
  <div class="app-shell">
    <aside class="sidebar">
      <a class="brand" href="#/dashboard">
        <span class="brand-mark">SM</span>
        <span>SMART<span class="brand-light">Maintenance</span></span>
      </a>

      <div class="sidebar-label">MENU UTAMA</div>

      <nav class="main-nav">
        <a
          href="#/dashboard"
          class={route.page === 'dashboard' ? 'active' : ''}
        >
          <span class="nav-icon">▦</span> Dashboard
        </a>

        <a
          href="#/assets"
          class={['assets', 'asset_detail', 'asset_new', 'asset_edit'].includes(route.page)
            ? 'active'
            : ''}
        >
          <span class="nav-icon">⬡</span>
          {currentUser?.role === 'Manager' ? 'Master Aset' : 'Aset Mesin'}
        </a>

        {#if currentUser?.role === 'Manager'}
          <a
            href="#/work-orders"
            class={route.page === 'work_orders' ? 'active' : ''}
          >
            <span class="nav-icon">▤</span> Work Order
          </a>

          <a
            href="#/users"
            class={route.page === 'users' ? 'active' : ''}
          >
            <span class="nav-icon">♙</span> Pengguna
          </a>
        {/if}
      </nav>

      <div class="sidebar-bottom">
        <div class="help-card">
          <span class="help-dot"></span>
          <strong>Sistem siap digunakan</strong>
          <span>Data maintenance tersimpan lokal.</span>
        </div>

        <div class="profile">
          <div class="avatar">
            {(currentUser?.full_name || 'U').charAt(0).toUpperCase()}
          </div>
          <div class="profile-info">
            <strong>{currentUser?.full_name}</strong>
            <span>{currentUser?.role}</span>
          </div>
          <button
            class="icon-button logout-button"
            type="button"
            title="Keluar"
            aria-label="Keluar"
            onclick={handleLogout}
          >
            ↪
          </button>
        </div>
      </div>
    </aside>

    <main class="main-content">
      <header class="topbar">
        <div class="breadcrumb">
          Workspace <span>/</span> {breadcrumb}
        </div>
        <div class="topbar-date">{todayStr}</div>
      </header>

      <section class="page-content">
        {#if route.page === 'dashboard'}
          <Dashboard {currentUser} />
        {:else if route.page === 'assets'}
          <Assets {currentUser} />
        {:else if route.page === 'asset_detail'}
          <AssetDetail assetId={route.id} {currentUser} />
        {:else if route.page === 'asset_new'}
          <AssetForm {currentUser} />
        {:else if route.page === 'asset_edit'}
          <AssetForm assetId={route.id} {currentUser} />
        {:else if route.page === 'work_orders'}
          <WorkOrders {currentUser} />
        {:else if route.page === 'report'}
          <MaintenanceReport ticketId={route.ticketId} {currentUser} />
        {:else if route.page === 'users'}
          <Users />
        {/if}
      </section>
    </main>
  </div>
{/if}
