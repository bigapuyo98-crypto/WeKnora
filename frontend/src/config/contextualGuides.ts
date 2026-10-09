// Section focus is shared by the real knowledge-base, agent and model settings flows.
export const KB_EDITOR_FOCUS_SECTION_EVENT = 'weknora:kb-editor-focus-section'
export const AGENT_EDITOR_FOCUS_SECTION_EVENT = 'weknora:agent-editor-focus-section'

export const focusKbEditorSection = (section: string) => {
  window.dispatchEvent(
    new CustomEvent(KB_EDITOR_FOCUS_SECTION_EVENT, { detail: { section } }),
  )
}

export const focusAgentEditorSection = (section: string) => {
  window.dispatchEvent(
    new CustomEvent(AGENT_EDITOR_FOCUS_SECTION_EVENT, { detail: { section } }),
  )
}
