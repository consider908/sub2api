//go:build unit

package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/httpclient"
	"github.com/stretchr/testify/require"
)

type kiroProfileResolverRepo struct {
	mockAccountRepoForGemini
	updated map[int64]map[string]any
}

func (r *kiroProfileResolverRepo) UpdateCredentials(_ context.Context, id int64, credentials map[string]any) error {
	if r.updated == nil {
		r.updated = map[int64]map[string]any{}
	}
	r.updated[id] = cloneCredentials(credentials)
	if acc, ok := r.accountsByID[id]; ok {
		acc.Credentials = cloneCredentials(credentials)
	}
	return nil
}

func TestKiroResolveAndPersistProfileArn_ResolvesAndPersistsMissingARN(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/", r.URL.Path)
		require.Equal(t, "AmazonCodeWhispererService.ListAvailableProfiles", r.Header.Get("X-Amz-Target"))
		require.Equal(t, "Bearer access-token", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"profiles":[{"arn":"arn:aws:codewhisperer:us-east-1:123456789012:profile/REAL"}]}`))
	}))
	defer server.Close()

	prevFactory := kiroHTTPClientFactory
	kiroHTTPClientFactory = func(_ httpclient.Options) (*http.Client, error) {
		return &http.Client{Transport: rewriteHostTransport(t, server)}, nil
	}
	defer func() { kiroHTTPClientFactory = prevFactory }()

	account := &Account{
		ID:       301,
		Platform: PlatformKiro,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"auth_method": "idc",
		},
	}
	repo := &kiroProfileResolverRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}

	arn := kiroResolveAndPersistProfileArn(context.Background(), repo, account, "access-token")
	require.Equal(t, "arn:aws:codewhisperer:us-east-1:123456789012:profile/REAL", arn)
	require.Equal(t, arn, account.GetCredential("profile_arn"))
	require.Equal(t, arn, repo.updated[account.ID]["profile_arn"])
}

func TestKiroResolveAndPersistProfileArn_SkipsExistingRealARN(t *testing.T) {
	account := &Account{
		ID:       302,
		Platform: PlatformKiro,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"profile_arn": "arn:aws:codewhisperer:us-east-1:123456789012:profile/EXISTING",
		},
	}
	repo := &kiroProfileResolverRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}

	arn := kiroResolveAndPersistProfileArn(context.Background(), repo, account, "access-token")
	require.Equal(t, "arn:aws:codewhisperer:us-east-1:123456789012:profile/EXISTING", arn)
	require.Nil(t, repo.updated)
}

func TestKiroResolveAndPersistProfileArn_RetriesAfterFailure(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := calls.Add(1)
		require.Equal(t, "AmazonCodeWhispererService.ListAvailableProfiles", r.Header.Get("X-Amz-Target"))
		if current == 1 {
			http.Error(w, `{"message":"temporary failure"}`, http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"profiles":[{"arn":"arn:aws:codewhisperer:us-east-1:123456789012:profile/RETRY"}]}`))
	}))
	defer server.Close()

	prevFactory := kiroHTTPClientFactory
	kiroHTTPClientFactory = func(_ httpclient.Options) (*http.Client, error) {
		return &http.Client{Transport: rewriteHostTransport(t, server)}, nil
	}
	defer func() { kiroHTTPClientFactory = prevFactory }()

	account := &Account{
		ID:       303,
		Platform: PlatformKiro,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"auth_method": "idc",
		},
	}
	repo := &kiroProfileResolverRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}

	first := kiroResolveAndPersistProfileArn(context.Background(), repo, account, "access-token")
	require.Empty(t, first)
	require.Equal(t, int32(1), calls.Load())
	require.Nil(t, repo.updated, fmt.Sprintf("unexpected persisted credentials after failure: %#v", repo.updated))

	second := kiroResolveAndPersistProfileArn(context.Background(), repo, account, "access-token")
	require.Equal(t, "arn:aws:codewhisperer:us-east-1:123456789012:profile/RETRY", second)
	require.Equal(t, int32(2), calls.Load())
	require.Equal(t, second, account.GetCredential("profile_arn"))
	require.Equal(t, second, repo.updated[account.ID]["profile_arn"])
}

func TestKiroOAuthBuildAccountCredentialsPreservesExternalIDPFields(t *testing.T) {
	service := &KiroOAuthService{}
	creds := service.BuildAccountCredentials(&KiroTokenInfo{
		AccessToken:   "access-token",
		RefreshToken:  "refresh-token",
		AuthMethod:    "external_idp",
		TokenEndpoint: "https://token.example.com",
		IssuerURL:     "https://issuer.example.com",
		Scopes:        "openid profile",
		ClientID:      "client-id",
		ClientSecret:  "client-secret",
		ProfileArn:    "arn:aws:codewhisperer:us-east-1:123456789012:profile/EXT",
	})

	require.Equal(t, "https://token.example.com", creds["token_endpoint"])
	require.Equal(t, "https://issuer.example.com", creds["issuer_url"])
	require.Equal(t, "openid profile", creds["scopes"])
	require.Equal(t, "client-id", creds["client_id"])
	require.Equal(t, "client-secret", creds["client_secret"])
	require.Equal(t, "refresh-token", creds["refresh_token"])
	require.Equal(t, "arn:aws:codewhisperer:us-east-1:123456789012:profile/EXT", creds["profile_arn"])
}
