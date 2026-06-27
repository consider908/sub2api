package admin

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/oauth"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type kiroRefreshAdminServiceStub struct {
	*stubAdminService
	lastUpdateCredentials map[string]any
	lastUpdateID          int64
}

func (s *kiroRefreshAdminServiceStub) UpdateAccount(ctx context.Context, id int64, input *service.UpdateAccountInput) (*service.Account, error) {
	s.lastUpdateID = id
	if input.Credentials != nil {
		s.lastUpdateCredentials = make(map[string]any, len(input.Credentials))
		for k, v := range input.Credentials {
			s.lastUpdateCredentials[k] = v
		}
	}
	return &service.Account{
		ID:          id,
		Platform:    service.PlatformKiro,
		Type:        service.AccountTypeOAuth,
		Credentials: s.lastUpdateCredentials,
		Status:      service.StatusActive,
	}, nil
}

type panicClaudeOAuthClient struct{}

func (panicClaudeOAuthClient) GetOrganizationUUID(context.Context, string, string) (string, error) {
	panic("claude oauth should not be called for kiro refresh")
}

func (panicClaudeOAuthClient) GetAuthorizationCode(context.Context, string, string, string, string, string, string) (string, error) {
	panic("claude oauth should not be called for kiro refresh")
}

func (panicClaudeOAuthClient) ExchangeCodeForToken(context.Context, string, string, string, string, bool) (*oauth.TokenResponse, error) {
	panic("claude oauth should not be called for kiro refresh")
}

func (panicClaudeOAuthClient) RefreshToken(context.Context, string, string) (*oauth.TokenResponse, error) {
	panic("claude oauth should not be called for kiro refresh")
}

func TestAccountHandlerRefreshSingleAccountKiroUsesKiroOAuthService(t *testing.T) {
	adminSvc := &kiroRefreshAdminServiceStub{stubAdminService: newStubAdminService()}
	var captured url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		captured, err = url.ParseQuery(string(body))
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","expires_in":3600}`))
	}))
	defer server.Close()

	handler := NewAccountHandler(
		adminSvc,
		service.NewOAuthService(nil, panicClaudeOAuthClient{}),
		nil,
		nil,
		nil,
		service.NewKiroOAuthService(nil),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	account := &service.Account{
		ID:       101,
		Platform: service.PlatformKiro,
		Type:     service.AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":   "old-access",
			"refresh_token":  "old-refresh",
			"auth_method":    "external_idp",
			"provider":       "ExternalIdp",
			"client_id":      "client-id",
			"token_endpoint": server.URL,
			"issuer_url":     "https://login.microsoftonline.com/test/v2.0",
			"scopes":         "api://app/.default offline_access",
			"custom_field":   "keep-me",
		},
	}

	updated, warning, err := handler.refreshSingleAccount(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, "", warning)
	require.NotNil(t, updated)
	require.Equal(t, account.ID, adminSvc.lastUpdateID)
	require.Equal(t, "refresh_token", captured.Get("grant_type"))
	require.Equal(t, "client-id", captured.Get("client_id"))
	require.Equal(t, "old-refresh", captured.Get("refresh_token"))
	require.Equal(t, "api://app/.default offline_access", captured.Get("scope"))
	require.Equal(t, "new-access", adminSvc.lastUpdateCredentials["access_token"])
	require.Equal(t, "old-refresh", adminSvc.lastUpdateCredentials["refresh_token"])
	require.Equal(t, server.URL, adminSvc.lastUpdateCredentials["token_endpoint"])
	require.Equal(t, "https://login.microsoftonline.com/test/v2.0", adminSvc.lastUpdateCredentials["issuer_url"])
	require.Equal(t, "api://app/.default offline_access", adminSvc.lastUpdateCredentials["scopes"])
	require.Equal(t, "keep-me", adminSvc.lastUpdateCredentials["custom_field"])
}

func TestAccountHandlerRefreshSingleAccountKiroWithoutServiceReturnsError(t *testing.T) {
	handler := NewAccountHandler(
		newStubAdminService(),
		service.NewOAuthService(nil, panicClaudeOAuthClient{}),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	_, _, err := handler.refreshSingleAccount(context.Background(), &service.Account{
		ID:          102,
		Platform:    service.PlatformKiro,
		Type:        service.AccountTypeOAuth,
		Credentials: map[string]any{"refresh_token": "refresh-token"},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "KIRO_OAUTH_SERVICE_UNAVAILABLE")
}
