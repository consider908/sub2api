//go:build unit

package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountUsageService_GetUsage_KiroDirectAPIKeySupported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/getUsageLimits", r.URL.Path)
		require.Equal(t, "Bearer kiro-api-key", r.Header.Get("Authorization"))
		require.Equal(t, "API_KEY", r.Header.Get("tokentype"))
		require.Empty(t, r.URL.Query().Get("profileArn"))
		_, _ = w.Write([]byte(`{
			"subscriptionInfo":{"subscriptionTitle":"KIRO PRO","type":"pro"},
			"usageBreakdownList":[{"resourceType":"CREDIT","currentUsage":10,"usageLimit":100}]
		}`))
	}))
	defer server.Close()

	prev := resolveKiroRuntimeEndpoint
	resolveKiroRuntimeEndpoint = func(string) string { return server.URL }
	defer func() { resolveKiroRuntimeEndpoint = prev }()

	account := &Account{
		ID:       9101,
		Platform: PlatformKiro,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "kiro-api-key",
		},
	}
	repo := &mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}}
	svc := NewAccountUsageService(repo, nil, nil, nil, nil, NewUsageCache(), nil, nil)

	usage, err := svc.GetUsage(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotNil(t, usage)
	require.NotNil(t, usage.KiroCredit)
	require.Equal(t, 10.0, usage.KiroCredit.CurrentUsage)
}

func TestAccountUsageService_GetPassiveUsage_KiroAPIKeyUnsupported(t *testing.T) {
	account := &Account{
		ID:       9102,
		Platform: PlatformKiro,
		Type:     AccountTypeAPIKey,
	}
	repo := &mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}}
	svc := NewAccountUsageService(repo, nil, nil, nil, nil, NewUsageCache(), nil, nil)

	usage, err := svc.GetPassiveUsage(context.Background(), account.ID)
	require.Nil(t, usage)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Kiro OAuth")
}
