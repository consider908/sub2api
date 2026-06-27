//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateProvider_AllowsGrokAndKiro(t *testing.T) {
	require.NoError(t, validateProvider(MonitorProviderGrok))
	require.NoError(t, validateProvider(MonitorProviderKiro))
}

func TestRunCheckForModel_GrokUsesOpenAICompatibleChatAdapter(t *testing.T) {
	h := &openAICaptureHandler{}
	endpoint := setupFakeOpenAI(t, h)

	res := runCheckForModel(context.Background(), MonitorProviderGrok, endpoint, "sk-grok", "grok-4.3", nil)

	require.Equal(t, MonitorStatusOperational, res.Status)
	require.Equal(t, providerOpenAIPath, h.lastPath)
	require.Equal(t, "grok-4.3", h.lastBody["model"])
	require.Equal(t, "Bearer sk-grok", h.lastHeaders.Get("Authorization"))
}
