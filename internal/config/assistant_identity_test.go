package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Cover the shipped prompt catalog, including intent overrides which can replace
// the normal system prompt at runtime. Technical paths are not assistant names.
func TestBuiltinAssistantIdentity(t *testing.T) {
	catalog, err := loadPromptTemplates("../../config")
	require.NoError(t, err)
	require.NotNil(t, catalog)

	groups := []struct {
		name      string
		templates []PromptTemplate
		count     int
	}{
		{"system", catalog.SystemPrompt, 6},
		{"intent", catalog.IntentPrompts, 7},
		{"agent", catalog.AgentSystemPrompt, 7},
	}
	for _, group := range groups {
		t.Run(group.name, func(t *testing.T) {
			require.Len(t, group.templates, group.count)
			for _, template := range group.templates {
				t.Run(template.ID, func(t *testing.T) {
					require.Contains(t, template.Content, "You are X,")
					require.Contains(t, template.Content, "developed by xiao")
					require.NotContains(t, template.Content, "WeKnora")
					require.NotContains(t, template.Content, "Tencent")
				})
			}
		})
	}
	t.Run("fallback", func(t *testing.T) {
		fallback := FindTemplateByID(catalog, "default_fallback_prompt")
		require.NotNil(t, fallback)
		require.Contains(t, fallback.Content, "You are X,")
		require.Contains(t, fallback.Content, "developed by xiao")
		require.NotContains(t, fallback.Content, "WeKnora")
		require.NotContains(t, fallback.Content, "Tencent")
	})
	t.Run("technical_paths_unchanged", func(t *testing.T) {
		installer := FindTemplateByID(catalog, "skill_installer")
		require.NotNil(t, installer)
		require.Contains(t, installer.Content, ".weknora/requirements.json")
	})
}
