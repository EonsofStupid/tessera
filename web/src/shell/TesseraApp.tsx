import {
  Link,
  Outlet,
  RouterProvider,
  createRootRoute,
  createRoute,
  createRouter,
  useRouterState,
} from '@tanstack/react-router'
import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { beginSignIn, clearSession, finishSignIn, loadEnvironment } from '../lib/auth'
import { getCapabilities, getOperatorActions, getOverview } from '../lib/api'
import type { CapabilityDiscovery, Environment, OperatorActionCatalog, Overview, ResourceState } from '../lib/contracts'
import { emitOperatorEvent } from '../lib/operatorEvents'

export interface TesseraAppProps {
  basePath?: string
}

type AppState = {
  capabilities: ResourceState<CapabilityDiscovery>
  overview: ResourceState<Overview>
  actions: ResourceState<OperatorActionCatalog>
  environment?: Environment
  refresh: () => void
  signIn: () => void
  signOut: () => void
}

const StateContext = createContext<AppState | null>(null)

function useAppState(): AppState {
  const value = useContext(StateContext)
  if (!value) throw new Error('Tessera application state is missing')
  return value
}

type IconName = 'home' | 'users' | 'apps' | 'federation' | 'access' | 'flows' | 'shield' | 'audit' | 'operations' | 'settings' | 'spark' | 'search' | 'bell' | 'chevron'

function Icon({ name, size = 18 }: { name: IconName; size?: number }) {
  const paths: Record<IconName, ReactNode> = {
    home: <><path d="M3 10.8 12 3l9 7.8"/><path d="M5.5 9.5V21h13V9.5M9 21v-6h6v6"/></>,
    users: <><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/></>,
    apps: <><rect x="3" y="3" width="7" height="7" rx="2"/><rect x="14" y="3" width="7" height="7" rx="2"/><rect x="3" y="14" width="7" height="7" rx="2"/><rect x="14" y="14" width="7" height="7" rx="2"/></>,
    federation: <><circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a15 15 0 0 1 0 18M12 3a15 15 0 0 0 0 18"/></>,
    access: <><path d="M21 2 9.6 13.4M15 3h6v6"/><path d="M12 5H5a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-7"/></>,
    flows: <><circle cx="5" cy="5" r="2"/><circle cx="19" cy="5" r="2"/><circle cx="12" cy="19" r="2"/><path d="M7 5h10M6 7l5 10M18 7l-5 10"/></>,
    shield: <><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10"/><path d="m9 12 2 2 4-5"/></>,
    audit: <><path d="M4 4h16v16H4z"/><path d="M8 9h8M8 13h8M8 17h5"/></>,
    operations: <><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/><circle cx="12" cy="12" r="3"/></>,
    settings: <><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06-2.83 2.83-.06-.06a1.7 1.7 0 0 0-1.88-.34 1.7 1.7 0 0 0-1.03 1.56V21h-4v-.09A1.7 1.7 0 0 0 9 19.36a1.7 1.7 0 0 0-1.88.34l-.06.06-2.83-2.83.06-.06A1.7 1.7 0 0 0 4.63 15 1.7 1.7 0 0 0 3.08 14H3v-4h.09A1.7 1.7 0 0 0 4.64 9a1.7 1.7 0 0 0-.34-1.88l-.06-.06 2.83-2.83.06.06A1.7 1.7 0 0 0 9 4.63 1.7 1.7 0 0 0 10 3.08V3h4v.09A1.7 1.7 0 0 0 15 4.64a1.7 1.7 0 0 0 1.88-.34l.06-.06 2.83 2.83-.06.06A1.7 1.7 0 0 0 19.37 9 1.7 1.7 0 0 0 20.92 10H21v4h-.09A1.7 1.7 0 0 0 19.4 15Z"/></>,
    spark: <path d="m12 2 1.4 5.6L19 9l-5.6 1.4L12 16l-1.4-5.6L5 9l5.6-1.4L12 2ZM19 16l.7 2.3L22 19l-2.3.7L19 22l-.7-2.3L16 19l2.3-.7L19 16Z"/>,
    search: <><circle cx="11" cy="11" r="7"/><path d="m20 20-4-4"/></>,
    bell: <><path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9"/><path d="M10 21h4"/></>,
    chevron: <path d="m9 18 6-6-6-6"/>,
  }
  return <svg aria-hidden="true" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">{paths[name]}</svg>
}

const nav = [
  { label: 'Start', items: [{ to: '/', label: 'Overview', icon: 'home' as IconName }] },
  { label: 'Manage', items: [
    { to: '/directory', label: 'Directory', icon: 'users' as IconName },
    { to: '/applications', label: 'Applications', icon: 'apps' as IconName },
    { to: '/federation', label: 'Federation', icon: 'federation' as IconName },
    { to: '/access', label: 'Access edges', icon: 'access' as IconName },
    { to: '/flows', label: 'Flows', icon: 'flows' as IconName },
  ] },
  { label: 'Protect', items: [
    { to: '/security', label: 'Security', icon: 'shield' as IconName },
    { to: '/audit', label: 'Audit', icon: 'audit' as IconName },
  ] },
  { label: 'Operate', items: [
    { to: '/operations', label: 'Operations', icon: 'operations' as IconName },
    { to: '/settings', label: 'Settings', icon: 'settings' as IconName },
  ] },
]

function StatusPill({ status }: { status: string }) {
  const label = status.replaceAll('_', ' ')
  return <span className={`status status--${status}`}>{label}</span>
}

function Brand() {
  return <div className="brand"><span className="brand-mark"><span/><span/><span/></span><span><strong>Tessera</strong><small>Identity control plane</small></span></div>
}

function AppShell() {
	const state = useAppState()
	const pathname = useRouterState({ select: routerState => routerState.location.pathname })
	const ready = state.overview.status === 'ready' ? state.overview.data?.readiness.status : 'attention'
  useEffect(() => {
    const routeID = `route.${pathname.replace(/^\/+|\/+$/g, '').replaceAll('/', '.') || 'overview'}`
    emitOperatorEvent({ routeId: routeID, eventType: 'route_opened' })
  }, [pathname])
  useEffect(() => {
    const handleClick = (event: MouseEvent) => {
      const target = event.target instanceof Element ? event.target.closest<HTMLElement>('[data-control-id]') : null
      if (!target?.dataset.controlId) return
      const routeID = `route.${window.location.pathname.replace('/ui/console', '').replace(/^\/+|\/+$/g, '').replaceAll('/', '.') || 'overview'}`
      emitOperatorEvent({ routeId: routeID, controlId: target.dataset.controlId, actionId: target.dataset.actionId, eventType: 'control_activated' })
    }
    document.addEventListener('click', handleClick)
    return () => document.removeEventListener('click', handleClick)
  }, [])
	return <div className="app-shell">
    <aside className="sidebar">
      <Brand />
      <div className="tenant-switcher"><span className="tenant-avatar">PC</span><span><small>Private cloud</small><strong>Primary community</strong></span><Icon name="chevron" size={15}/></div>
      <nav aria-label="Primary navigation">
		{nav.map(group => <div className="nav-group" key={group.label}><p>{group.label}</p>{group.items.map(item => <Link key={item.to} to={item.to} data-control-id={`control.navigation_${item.label.toLowerCase().replaceAll(' ', '_')}`} activeProps={{ className: 'active' }} activeOptions={{ exact: item.to === '/' }}><Icon name={item.icon}/><span>{item.label}</span></Link>)}</div>)}
      </nav>
		<div className="sidebar-footer"><div className="trust-mini"><span className={`pulse pulse--${ready}`}/><span><strong>{ready === 'ready' ? 'Trust fabric ready' : 'Verification required'}</strong><small>{ready === 'ready' ? 'All required checks passed' : 'Sign in to inspect evidence'}</small></span></div><button type="button" data-control-id="control.session_clear" onClick={state.signOut}>Clear local session</button></div>
    </aside>
    <main>
		<header className="topbar"><button className="mobile-brand" type="button" data-control-id="control.mobile_navigation"><Brand /></button><label className="search"><Icon name="search" size={17}/><input aria-label="Search Tessera" data-control-id="control.search" placeholder="Search people, apps, policies…"/><kbd>⌘ K</kbd></label><div className="top-actions"><button className="icon-button" type="button" data-control-id="control.notifications" aria-label="Notifications"><Icon name="bell"/></button><button className="operator" type="button" data-control-id="control.operator_menu"><span>TE</span><span><strong>Tessera owner</strong><small>Private operator</small></span></button></div></header>
      <div className="page"><Outlet /></div>
    </main>
  </div>
}

function PageHeading({ eyebrow, title, copy, action }: { eyebrow: string; title: string; copy: string; action?: string }) {
	return <div className="page-heading"><div><span className="eyebrow">{eyebrow}</span><h1>{title}</h1><p>{copy}</p></div>{action && <button className="primary-button" type="button" data-control-id="control.primary_action" disabled><Icon name="spark" size={16}/>{action}<span className="preview-tag">Preview</span></button>}</div>
}

const fallbackCapabilities = [
  ['upstream_oidc', 'OIDC federation'], ['upstream_saml', 'SAML federation'], ['ldap_outbound', 'LDAP outbound'],
  ['ldap_inbound', 'LDAP inbound'], ['forward_auth', 'Forward auth'], ['identity_aware_proxy', 'Identity-aware proxy'],
  ['visual_flow_engine', 'Visual flow engine'], ['vaultix_secret_custody', 'Vaultix custody'], ['analytics_olap', 'ClickHouse analytics'],
] as const

function OverviewPage() {
  const state = useAppState()
  const overview = state.overview.data
  const facts = state.capabilities.data?.capabilities
	const capabilities = fallbackCapabilities.map(([id, label]) => ({ id, label, fact: facts?.find(item => item.id === id) }))
	const actions = state.actions.data?.actions ?? []
	const enabledActions = actions.filter(action => action.exposure === 'enabled').length
  const lenses = overview?.lenses ?? [
    { id: 'infrastructure', label: 'Infrastructure', value: 0, unit: 'attachments', detail: 'Identity attachments become visible after sign-in.', status: 'degraded' },
    { id: 'ai', label: 'AI', value: 0, unit: 'agent seats', detail: 'Agent identities and delegated scopes stay explicit.', status: 'degraded' },
    { id: 'customers', label: 'Customers', value: 0, unit: 'human seats', detail: 'Tenant-scoped people in your private community.', status: 'degraded' },
  ]
  return <>
    <PageHeading eyebrow="Command center" title="Good evening, operator." copy="See who can reach what, why they can reach it, and whether every trust path is proven." />
		{state.overview.status !== 'ready' && <div className="callout"><span className="callout-icon"><Icon name="shield"/></span><div><strong>{state.overview.status === 'authentication_required' ? 'Connect your Tessera session' : 'Live identity facts are unavailable'}</strong><p>{state.overview.error?.message ?? 'Sign in to load tenant counts, readiness evidence, and deployment facts. No placeholder is treated as operational.'}</p></div><button type="button" data-control-id={state.overview.status === 'authentication_required' ? 'control.sign_in' : 'control.retry_overview'} onClick={state.overview.status === 'authentication_required' ? state.signIn : state.refresh}>{state.overview.status === 'authentication_required' ? 'Sign in' : 'Retry'}</button></div>}
    <section className="hero-grid">
      <article className="trust-card"><div className="trust-orbit"><span className="orbit orbit-one"/><span className="orbit orbit-two"/><span className="orbit-core"><Icon name="shield" size={32}/></span></div><div><span className="eyebrow">Trust posture</span><h2>{overview?.readiness.status === 'ready' ? 'Identity fabric is ready' : 'Evidence is required'}</h2><p>{overview?.readiness.status === 'ready' ? 'Signing, policy and flow checks currently meet the declared readiness policy.' : 'Tessera remains fail-closed until live readiness and conformance evidence can be read.'}</p><div className="trust-stats"><span><strong>{overview?.readiness.signing_keys ?? '—'}</strong> signing keys</span><span><strong>{overview?.readiness.flows ?? '—'}</strong> active flows</span><span><strong>{state.capabilities.data?.bundle_manifest_digest ? 'Bound' : 'Unattested'}</strong> bundle</span></div></div></article>
		<article className="guide-card"><span className="guide-icon"><Icon name="spark" size={22}/></span><span className="eyebrow">Guided next step</span><h2>{state.overview.status === 'ready' ? 'Review recovery posture' : 'Establish operator access'}</h2><p>{state.overview.status === 'ready' ? 'Verify that owner recovery remains independent before the next upgrade.' : 'Authenticate as the first owner, then Tessera will build a reviewed plan from observed state.'}</p><button type="button" data-control-id="control.guide_open" disabled>Open guided plan <Icon name="chevron" size={15}/></button></article>
    </section>
    <section><div className="section-heading"><div><span className="eyebrow">Identity lenses</span><h2>Your private ecosystem, separated by purpose</h2></div><span className="freshness">{overview ? `Observed ${new Date(overview.observed_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}` : 'Awaiting live projection'}</span></div><div className="lens-grid">{lenses.map((lens, index) => <article className={`lens-card lens-${index}`} key={lens.id}><div className="lens-top"><span className="lens-symbol">{index === 0 ? '⌁' : index === 1 ? '✦' : '◌'}</span><StatusPill status={lens.status}/></div><strong className="lens-value">{lens.value}</strong><h3>{lens.label}</h3><p>{lens.unit}</p><small>{lens.detail}</small></article>)}</div></section>
		<section><div className="section-heading"><div><span className="eyebrow">Capability proof</span><h2>Federation and access edges</h2></div><Link to="/federation" data-control-id="control.federation_review">Review federation <Icon name="chevron" size={14}/></Link></div><div className="capability-list">{capabilities.map(({ id, label, fact }) => <div className="capability-row" key={id}><span className="capability-glyph"><Icon name={id.includes('ldap') ? 'users' : id.includes('proxy') || id.includes('auth') ? 'access' : id.includes('flow') ? 'flows' : id.includes('vault') ? 'shield' : 'federation'} size={17}/></span><span><strong>{label}</strong><small>{fact?.reason?.replaceAll('_', ' ') ?? 'Live discovery not loaded'}</small></span><StatusPill status={fact?.status ?? 'preview'}/></div>)}</div></section>
		<section className="operator-plane"><div><span className="eyebrow">Human + AI operator plane</span><h2>Clyffy works from the same reviewed controls you do.</h2><p>Stable action schemas expose consequence, permission, assurance, seed suggestions and verification. No specialist gets a hidden administrative route.</p></div><div className="operator-flow"><span><small>Human intent</small><strong>Review</strong></span><i>→</i><span><small>Clyffy specialist</small><strong>Plan</strong></span><i>→</i><span><small>PostgreSQL truth</small><strong>Verify</strong></span></div><div className="operator-score"><strong>{enabledActions}</strong><span>executable actions</span><small>{actions.length ? `${actions.length} discovered · unavailable actions stay disabled` : 'Sign in to discover the authorized action catalog'}</small></div></section>
	</>
}

type Feature = { title: string; copy: string; status?: string; metric?: string }
const pageContent: Record<string, { eyebrow: string; title: string; copy: string; action: string; features: Feature[] }> = {
  directory: { eyebrow: 'Directory', title: 'People, groups, and machine identities', copy: 'One place to understand every identity and the tenant boundary it belongs to.', action: 'Add identity', features: [
    { title: 'Human identities', copy: 'Owners, members, guests and lifecycle state.', metric: 'Tenant scoped' }, { title: 'Service identities', copy: 'Workloads with explicit credentials and rotation policy.', status: 'preview' }, { title: 'Groups', copy: 'Reusable membership without hidden privilege inheritance.', status: 'preview' },
  ] },
  applications: { eyebrow: 'Applications', title: 'Connect applications without guesswork', copy: 'Guided protocol choices, redirect validation, keys, consent and token policy.', action: 'Connect application', features: [
    { title: 'Web and native apps', copy: 'Authorization code with PKCE and reviewed redirects.', status: 'preview' }, { title: 'Machine-to-machine', copy: 'Least-privilege service identities and asymmetric trust.', status: 'preview' }, { title: 'SAML service providers', copy: 'Metadata, signing and assertion policy in one guided path.', status: 'preview' },
  ] },
  federation: { eyebrow: 'Federation', title: 'Bring every trusted identity source together', copy: 'Upstream identity providers and downstream relying parties with explicit trust evidence.', action: 'Add provider', features: [
    { title: 'OpenID Connect', copy: 'Discovery, claims, account linking and logout conformance.', status: 'preview' }, { title: 'SAML 2.0', copy: 'Metadata, encrypted assertions and certificate lifecycle.', status: 'preview' }, { title: 'LDAP directories', copy: 'Outbound synchronization and inbound bind are promoted independently.', status: 'preview' },
  ] },
  access: { eyebrow: 'Access edges', title: 'Protect applications at the edge', copy: 'Make identity-aware access available to applications that cannot speak modern federation.', action: 'Create access edge', features: [
    { title: 'Forward auth', copy: 'Header, cookie, websocket, logout and bypass-resistance checks.', status: 'preview' }, { title: 'Identity-aware proxy', copy: 'Policy enforcement with Zuul mesh integration when installed.', status: 'preview' }, { title: 'LDAP gateway', copy: 'Inbound and outbound directory paths fail closed independently.', status: 'preview' },
  ] },
  flows: { eyebrow: 'Flows', title: 'Design policy as a visible journey', copy: 'A guided visual engine for authentication, enrollment, recovery and conditional access.', action: 'Create flow', features: [
    { title: 'Authentication', copy: 'Passwordless, MFA and federated choices as typed nodes.', status: 'preview' }, { title: 'Enrollment', copy: 'Invite-only journeys with reviewable conditions.', status: 'preview' }, { title: 'Recovery', copy: 'Independent owner recovery that cannot silently weaken policy.', status: 'preview' },
  ] },
  security: { eyebrow: 'Security', title: 'Make trust posture explain itself', copy: 'Keys, authenticators, sessions, custody and risk signals with a safe remedy for every finding.', action: 'Run posture review', features: [
    { title: 'Signing keys', copy: 'Asymmetric keys, rotation windows and verifier overlap.', status: 'preview' }, { title: 'Vaultix custody', copy: 'Protected references only; secret values never enter browser state.', status: 'preview' }, { title: 'Session policy', copy: 'Assurance, lifetime and revocation controls.', status: 'preview' },
  ] },
  audit: { eyebrow: 'Audit', title: 'Trace every identity decision', copy: 'Transactional evidence stays authoritative while ClickHouse serves rebuildable OLAP views.', action: 'Create export', features: [
    { title: 'Security events', copy: 'Actor, target, tenant, decision and correlation context.', status: 'preview' }, { title: 'Analytics projection', copy: 'Asynchronous PostgreSQL outbox to ClickHouse; auth never waits.', status: 'preview' }, { title: 'Exports and retention', copy: 'Tenant-visible evidence with explicit retention policy.', status: 'preview' },
  ] },
  operations: { eyebrow: 'Operations', title: 'Operate Tessera with reviewed changes', copy: 'Preflight, plan, apply and verify installation, backup, restore, upgrades and rotations.', action: 'Plan operation', features: [
    { title: 'Deployment lifecycle', copy: 'A server-owned state backed by contributing checks.', status: 'preview' }, { title: 'Backup and restore', copy: 'Restoration is complete only after identity outcomes verify.', status: 'preview' }, { title: 'Tessera operator', copy: 'A separate least-privilege Podman lifecycle component.', status: 'preview' },
  ] },
  settings: { eyebrow: 'Settings', title: 'Shape this Tessera deployment', copy: 'Tenant boundaries, public issuer, branding, notification and optional integrations.', action: 'Review changes', features: [
    { title: 'Deployment profile', copy: 'Dedicated runtime or invite-only community isolation.', metric: 'Explicit' }, { title: 'Public issuer', copy: 'A migration-controlled trust root, not a restart option.', status: 'preview' }, { title: 'Host integrations', copy: 'Optional adapters remain outside Tessera’s standalone boundary.', status: 'preview' },
  ] },
}

function FeaturePage({ page }: { page: keyof typeof pageContent }) {
  const content = pageContent[page]
	return <><PageHeading eyebrow={content.eyebrow} title={content.title} copy={content.copy} action={content.action}/><div className="feature-grid">{content.features.map((feature, index) => <article className="feature-card" key={feature.title}><span className="feature-index">0{index + 1}</span><div>{feature.metric ? <span className="metric-tag">{feature.metric}</span> : <StatusPill status={feature.status ?? 'preview'}/>}<h2>{feature.title}</h2><p>{feature.copy}</p></div><button type="button" data-control-id={`control.feature_${page}_${index + 1}`} aria-label={`Open ${feature.title}`} disabled><Icon name="chevron"/></button></article>)}</div><article className="empty-workbench"><span><Icon name="spark" size={24}/></span><div><h2>Guided workbench</h2><p>This section becomes actionable as its server capability earns passing, bundle-bound conformance evidence. Until then, it stays visible for review and disabled for safety.</p></div><StatusPill status="preview"/></article></>
}

function CallbackPage() {
  return <div className="callback"><Brand/><span className="loader"/><h1>Establishing your Tessera session</h1><p>Validating the authorization response and loading your identity boundary.</p></div>
}

function StateProvider({ children }: { children: ReactNode }) {
  const [refreshKey, setRefreshKey] = useState(0)
  const [environment, setEnvironment] = useState<Environment>()
  const [capabilities, setCapabilities] = useState<ResourceState<CapabilityDiscovery>>({ status: 'loading' })
  const [overview, setOverview] = useState<ResourceState<Overview>>({ status: 'loading' })
  const [actions, setActions] = useState<ResourceState<OperatorActionCatalog>>({ status: 'loading' })

  useEffect(() => {
    let active = true
    void loadEnvironment().then(async env => {
      if (!active) return
      setEnvironment(env)
      if (window.location.pathname.endsWith('/auth/callback')) {
        await finishSignIn(env)
      }
		const [nextCapabilities, nextOverview, nextActions] = await Promise.all([getCapabilities(), getOverview(), getOperatorActions()])
      if (active) {
        setCapabilities(nextCapabilities)
			setOverview(nextOverview)
			setActions(nextActions)
      }
    }).catch(() => {
      if (active) {
        setCapabilities({ status: 'authentication_required' })
		setOverview({ status: 'authentication_required' })
		setActions({ status: 'authentication_required' })
      }
    })
    return () => { active = false }
  }, [refreshKey])

  const value = useMemo<AppState>(() => ({
	capabilities,
	overview,
	actions,
    environment,
    refresh: () => setRefreshKey(value => value + 1),
    signIn: () => { if (environment) void beginSignIn(environment) },
    signOut: () => { clearSession(); setRefreshKey(value => value + 1) },
	}), [capabilities, overview, actions, environment])

  return <StateContext.Provider value={value}>{children}</StateContext.Provider>
}

const rootRoute = createRootRoute({ component: AppShell })
const overviewRoute = createRoute({ getParentRoute: () => rootRoute, path: '/', component: OverviewPage })
const routes = Object.keys(pageContent).map(path => createRoute({ getParentRoute: () => rootRoute, path: `/${path}`, component: () => <FeaturePage page={path}/> }))
const callbackRoute = createRoute({ getParentRoute: () => rootRoute, path: '/auth/callback', component: CallbackPage })
const routeTree = rootRoute.addChildren([overviewRoute, ...routes, callbackRoute])

export function TesseraApp({ basePath = '/ui/console' }: TesseraAppProps) {
  const router = useMemo(() => createRouter({ routeTree, basepath: basePath, defaultPreload: 'intent' }), [basePath])
  return <StateProvider><RouterProvider router={router}/></StateProvider>
}

declare module '@tanstack/react-router' {
  interface Register { router: ReturnType<typeof createRouter> }
}
