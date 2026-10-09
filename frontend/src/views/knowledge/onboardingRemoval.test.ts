import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { LOCALE_BUNDLES, collectLocaleKeys, getLocaleValueAtPath } from '../../i18n/localeKeyAudit'

test('all locales drop tutorial and help-entry copy but keep the model prerequisite warning', () => {
  const warningKey = 'contextualGuide.tenantModels.needChatModelFirst'
  for (const [locale, bundle] of Object.entries(LOCALE_BUNDLES)) {
    const stale = [...collectLocaleKeys(bundle)].filter(key =>
      key.startsWith('newUserGuide.') ||
      (key.startsWith('contextualGuide.') && key !== warningKey) ||
      key === 'general.helpAndDocs' || key === 'commandPalette.quick.productTour',
    )
    assert.deepEqual(stale, [], locale)
    assert.equal(typeof getLocaleValueAtPath(bundle, warningKey), 'string', locale)
  }
})

const read = (path: string) => readFileSync(new URL(path, import.meta.url), 'utf8')

// Static owner guards complement authenticated UI checks. They explicitly
// protect real editor, model-readiness and list-refresh behavior.
test('business pages do not mount tutorials or write tutorial progress', () => {
  const pages = [
    'KnowledgeBaseList.vue', 'KnowledgeBase.vue', 'KnowledgeBaseEditorModal.vue',
    '../agent/AgentList.vue', '../agent/AgentEditorModal.vue', '../creatChat/creatChat.vue',
  ]
  for (const page of pages) {
    assert.doesNotMatch(read(page), /ContextualGuide|TenantModelsGuide|markContextualGuideDone|isContextualGuideDone/, page)
  }
})

test('removing tutorials preserves editor creation and list refresh paths', () => {
  const kbEditor = read('KnowledgeBaseEditorModal.vue')
  const agentEditor = read('../agent/AgentEditorModal.vue')
  const kbList = read('KnowledgeBaseList.vue')
  const agentList = read('../agent/AgentList.vue')
  assert.match(kbEditor, /await createKnowledgeBase\(data\)/)
  assert.match(kbEditor, /emit\('success', createdKbId\)/)
  assert.match(agentEditor, /await createAgent\(payload\)/)
  assert.match(agentEditor, /modalShell\.markClean\(\)/)
  assert.match(kbList, /chatResources\.invalidateKnowledgeBaseDetail\(kbId\)/)
  assert.match(kbList, /fetchList\(true\)/)
  assert.match(kbList, /uiStore\.openCreateKB\('document', initialSection\)/)
  assert.match(agentList, /if \(!isReadyForAgent\.value\)/)
  assert.match(agentList, /MessagePlugin\.warning/)
  assert.match(agentList, /uiStore\.openSettings\('models'\)/)
})

test('editor focus events remain available independently of tutorials', () => {
  const config = read('../../config/contextualGuides.ts')
  for (const event of ['weknora:kb-editor-focus-section', 'weknora:agent-editor-focus-section']) {
    assert.ok(config.includes(event))
  }
  assert.match(read('KnowledgeBaseEditorModal.vue'), /window\.addEventListener\(KB_EDITOR_FOCUS_SECTION_EVENT/)
  assert.match(read('../agent/AgentEditorModal.vue'), /window\.addEventListener\(AGENT_EDITOR_FOCUS_SECTION_EVENT/)
  assert.match(read('../settings/ModelSettings.vue'), /focusKbEditorSection\(knowledgeBaseSection\)/)
  assert.match(read('../agent/AgentList.vue'), /focusAgentEditorSection\(requestedSection\)/)
})
