import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { parse } from '@vue/compiler-sfc'

// Source-level presentation guard; real login/register interactions are also
// verified in the local browser. This does not exercise authentication APIs.
const frontend = new URL('../../../', import.meta.url)
const read = (path: string) => readFileSync(new URL(path, frontend), 'utf8')
const template = (path: string) => {
  const { descriptor, errors } = parse(read(path))
  assert.deepEqual(errors, [])
  assert.ok(descriptor.template)
  return descriptor.template.content
}

test('authentication screens do not expose upstream branding or promotional panels', () => {
  const auth = template('src/views/auth/Login.vue')
  assert.doesNotMatch(auth, /WeKnora|weknora\.(png|weixin\.qq\.com)|github\.com\/Tencent/i)
  assert.doesNotMatch(auth, /['"]platform\.|auth\.(subtitle|firstTime|registerSubtitle)|<swiper(?:\s|-)/)
})

test('shared navigation omits the upstream logo and GitHub promotion without losing controls', () => {
  const sidebar = template('src/components/menu.vue')
  const userMenu = template('src/components/UserMenu.vue')
  assert.doesNotMatch(sidebar, /weknora\.png/i)
  assert.doesNotMatch(userMenu, /common\.github|logo-github|github\.com\/Tencent/i)
  for (const control of ['commandPaletteStore.openPalette', 'uiStore.toggleSidebar', 'TenantSelector']) {
    assert.ok(sidebar.includes(control), `Missing sidebar control: ${control}`)
  }
  for (const control of ['handleSettings', 'handleLogout', 'handleSystemAdmin']) {
    assert.ok(userMenu.includes(control), `Missing account control: ${control}`)
  }
})

test('browser entry heads have no upstream branding or legacy brand icons', () => {
  for (const path of ['index.html', 'embed.html']) {
    const head = read(path).split('</head>')[0]!
    assert.doesNotMatch(head, /WeKnora|public\/favicon\.ico/i, path)
    assert.match(head, /<title>[^<]+<\/title>/, path)
    // Explicit empty favicon prevents the browser's automatic /favicon.ico lookup.
    assert.match(head, /<link[^>]+rel="icon"[^>]+href="data:,"/, path)
  }
  assert.match(read('index.html'), /localStorage\.getItem\('WeKnora_theme'\)/)
})

test('authentication forms retain submit, registration, invitation, OIDC and language controls', () => {
  const auth = template('src/views/auth/Login.vue')
  for (const contract of [
    '@submit="handleLogin"', '@submit="handleRegister"', '@click="toggleMode"',
    'registrationEnabled', 'inviteLookup', 'inviteLookupError',
    '@click="handleOIDCLogin"', 'v-if="oidcEnabled"', '@click="toggleLanguageMenu"',
    '@click="selectLanguage(lang.value)"', 'autocomplete="current-password"',
    'autocomplete="new-password"', ':rules="formRules"', ':rules="registerRules"',
  ]) assert.ok(auth.includes(contract), `Missing functional control: ${contract}`)
})
