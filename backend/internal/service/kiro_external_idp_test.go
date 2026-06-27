package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKiroOAuthServiceBuildAccountCredentialsIncludesExternalIDPFields(t *testing.T) {
	svc := NewKiroOAuthService(nil)

	creds := svc.BuildAccountCredentials(&KiroTokenInfo{
		AccessToken:   "access-token",
		RefreshToken:  "refresh-token",
		AuthMethod:    "external_idp",
		Provider:      "ExternalIdp",
		ClientID:      "client-id",
		TokenEndpoint: "https://login.microsoftonline.com/test/oauth2/v2.0/token",
		IssuerURL:     "https://login.microsoftonline.com/test/v2.0",
		Scopes:        "api://app/.default offline_access",
	})

	require.Equal(t, "https://login.microsoftonline.com/test/oauth2/v2.0/token", creds["token_endpoint"])
	require.Equal(t, "https://login.microsoftonline.com/test/v2.0", creds["issuer_url"])
	require.Equal(t, "api://app/.default offline_access", creds["scopes"])
}

func TestKiroOAuthServiceRefreshTokenExternalIDPUsesStoredFieldsAndBackfillsMissingResponseFields(t *testing.T) {
	svc := NewKiroOAuthService(nil)
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

	info, err := svc.RefreshToken(context.Background(), &KiroRefreshTokenInput{
		RefreshToken:  "refresh-token",
		AuthMethod:    "external_idp",
		Provider:      "ExternalIdp",
		ClientID:      "client-id",
		TokenEndpoint: server.URL,
		IssuerURL:     "https://login.microsoftonline.com/test/v2.0",
		Scopes:        "api://app/.default offline_access",
	})
	require.NoError(t, err)
	require.Equal(t, "refresh_token", captured.Get("grant_type"))
	require.Equal(t, "client-id", captured.Get("client_id"))
	require.Equal(t, "refresh-token", captured.Get("refresh_token"))
	require.Equal(t, "api://app/.default offline_access", captured.Get("scope"))
	require.Equal(t, "new-access", info.AccessToken)
	require.Equal(t, "refresh-token", info.RefreshToken)
	require.Equal(t, server.URL, info.TokenEndpoint)
	require.Equal(t, "https://login.microsoftonline.com/test/v2.0", info.IssuerURL)
	require.Equal(t, "api://app/.default offline_access", info.Scopes)
}
