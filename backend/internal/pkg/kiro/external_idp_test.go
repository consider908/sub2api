package kiro

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseImportedTokenExternalIDPPreservesFields(t *testing.T) {
	token, err := ParseImportedToken(`{
		"accessToken":"access-token",
		"refreshToken":"refresh-token",
		"authMethod":" external_idp ",
		"provider":"",
		"clientId":"client-id",
		"tokenEndpoint":"https://login.microsoftonline.com/test/oauth2/v2.0/token",
		"issuerUrl":"https://login.microsoftonline.com/test/v2.0",
		"scopes":"api://app/.default offline_access"
	}`, "")
	require.NoError(t, err)

	require.Equal(t, "external_idp", token.AuthMethod)
	require.Equal(t, "ExternalIdp", token.Provider)
	require.Equal(t, "client-id", token.ClientID)
	require.Equal(t, "https://login.microsoftonline.com/test/oauth2/v2.0/token", token.TokenEndpoint)
	require.Equal(t, "https://login.microsoftonline.com/test/v2.0", token.IssuerURL)
	require.Equal(t, "api://app/.default offline_access", token.Scopes)
}

func TestRefreshExternalIDPTokenUsesFormURLEncodedRefreshGrant(t *testing.T) {
	var capturedContentType string
	var captured url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		capturedContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		captured, err = url.ParseQuery(string(body))
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","refresh_token":"new-refresh","expires_in":7200}`))
	}))
	defer server.Close()

	token, err := RefreshExternalIDPToken(
		context.Background(),
		"",
		"client-id",
		"old-refresh",
		server.URL,
		"api://app/.default offline_access",
	)
	require.NoError(t, err)

	require.Equal(t, "application/x-www-form-urlencoded", capturedContentType)
	require.Equal(t, "refresh_token", captured.Get("grant_type"))
	require.Equal(t, "client-id", captured.Get("client_id"))
	require.Equal(t, "old-refresh", captured.Get("refresh_token"))
	require.Equal(t, "api://app/.default offline_access", captured.Get("scope"))
	require.Equal(t, "new-access", token.AccessToken)
	require.Equal(t, "new-refresh", token.RefreshToken)
	require.Equal(t, "external_idp", token.AuthMethod)
	require.Equal(t, "ExternalIdp", token.Provider)
	require.Equal(t, server.URL, token.TokenEndpoint)
	require.Equal(t, "api://app/.default offline_access", token.Scopes)
}

func TestRefreshExternalIDPTokenKeepsOldRefreshTokenWhenResponseOmitsIt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","expires_in":3600}`))
	}))
	defer server.Close()

	token, err := RefreshExternalIDPToken(context.Background(), "", "client-id", "old-refresh", server.URL, "offline_access")
	require.NoError(t, err)
	require.Equal(t, "old-refresh", token.RefreshToken)
}

func TestRefreshExternalIDPTokenInvalidGrantReturnsTypedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"AADSTS70000"}`))
	}))
	defer server.Close()

	_, err := RefreshExternalIDPToken(context.Background(), "", "client-id", "revoked-refresh", server.URL, "offline_access")
	require.Error(t, err)

	var invalid *RefreshTokenInvalidError
	require.True(t, errors.As(err, &invalid))
	require.True(t, strings.Contains(strings.ToLower(invalid.Body), "invalid_grant"))
}
