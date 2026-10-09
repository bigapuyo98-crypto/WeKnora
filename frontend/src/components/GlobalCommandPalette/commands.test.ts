import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { after, before, test } from 'node:test'
import { fileURLToPath } from 'node:url'
import { createServer, type ViteDevServer } from 'vite'
import type { Router } from 'vue-router'
import type { Composer } from 'vue-i18n'
let server: ViteDevServer
let buildCommands: typeof import('./commands')['buildCommands']
before(async () => {
  server = await createServer({
    configFile: false, optimizeDeps: { noDiscovery: true, entries: [] },
    resolve: { alias: { '@': fileURLToPath(new URL('../../', import.meta.url)) } },
    server: { middlewareMode: true, hmr: false }, appType: 'custom',
  })
  buildCommands = (await server.ssrLoadModule('/src/components/GlobalCommandPalette/commands.ts')).buildCommands
})
after(async () => { await server?.close() })

const source = (path: string) => readFileSync(new URL(path, import.meta.url), 'utf8')

test('platform and account menu no longer offer welcome tours or the documentation entry', () => {
  const platform = source('../../views/platform/index.vue')
  const userMenu = source('../UserMenu.vue')
  assert.doesNotMatch(platform, /NewUserGuide/)
  assert.doesNotMatch(userMenu, /openNewUserGuide|reopenGuide|newUserGuide|openDocs|helpAndDocs/)
  for (const capability of ['GlobalCommandPalette', 'UploadTasksPanel', 'GlobalInvitationBell', '<Settings']) {
    assert.ok(platform.includes(capability), `Missing platform capability: ${capability}`)
  }
  for (const capability of ['handleSettings', 'handleLogout', 'handleSystemAdmin']) {
    assert.ok(userMenu.includes(capability), `Missing account capability: ${capability}`)
  }
})

test('command palette keeps business navigation without a product-tour action', async () => {
  const visited: string[] = []
  let closed = 0
  const commands = buildCommands({
    router: { push: (path: string) => { visited.push(path) } } as unknown as Router,
    t: ((key: string) => key) as Composer['t'],
    close: () => { closed++ },
  })
  assert.deepEqual(commands.map(command => command.id), [
    'new-chat', 'open-kb-list', 'open-agents', 'open-organizations', 'open-settings',
  ])
  for (const command of commands) await command.run()
  assert.deepEqual(visited, [
    '/platform/creatChat', '/platform/knowledge-bases', '/platform/agents',
    '/platform/organizations', '/platform/settings',
  ])
  assert.equal(closed, 5)
})
