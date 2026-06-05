//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type updateServiceCacheStub struct {
	data string
}

func (s *updateServiceCacheStub) GetUpdateInfo(context.Context) (string, error) {
	if s.data == "" {
		return "", errors.New("cache miss")
	}
	return s.data, nil
}

func (s *updateServiceCacheStub) SetUpdateInfo(_ context.Context, data string, _ time.Duration) error {
	s.data = data
	return nil
}

type updateServiceGitHubClientStub struct {
	release *GitHubRelease
	repo    string
	err     error
}

func (s *updateServiceGitHubClientStub) FetchLatestRelease(_ context.Context, repo string) (*GitHubRelease, error) {
	s.repo = repo
	if s.err != nil {
		return nil, s.err
	}
	return s.release, nil
}

func (s *updateServiceGitHubClientStub) DownloadFile(context.Context, string, string, int64) error {
	panic("DownloadFile should not be called when no update is available")
}

func (s *updateServiceGitHubClientStub) FetchChecksumFile(context.Context, string) ([]byte, error) {
	panic("FetchChecksumFile should not be called when no update is available")
}

func TestUpdateServicePerformUpdateNoUpdateReturnsSentinel(t *testing.T) {
	svc := NewUpdateService(
		&updateServiceCacheStub{},
		&updateServiceGitHubClientStub{
			release: &GitHubRelease{
				TagName: "v0.1.132",
				Name:    "v0.1.132",
			},
		},
		"0.1.132",
		"release",
	)

	err := svc.PerformUpdate(context.Background())

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoUpdateAvailable))
	require.ErrorIs(t, err, ErrNoUpdateAvailable)
}

func TestUpdateServiceVersionComparisonSupportsLocalRevision(t *testing.T) {
	require.Negative(t, compareVersions("0.1.133.1", "0.1.134.1"))
	require.Zero(t, compareVersions("0.1.134", "0.1.134.0"))
	require.Positive(t, compareVersions("0.1.133.2", "0.1.133.1"))
	require.Positive(t, compareVersions("0.1.134.1", "0.1.133.99"))
}

func TestUpdateServiceVersionComparisonHandlesInvalidVersions(t *testing.T) {
	require.Zero(t, compareVersions("0.1.134.1", "bad-version"))
	require.Negative(t, compareVersions("bad-version", "0.1.134.1"))
	require.Zero(t, compareVersions("bad-version", "also-bad"))
}

func TestParseProductVersionBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		normalized string
		parts      [4]int
		ok         bool
	}{
		{
			name:       "upstream three part",
			version:    "v0.1.134",
			normalized: "0.1.134",
			parts:      [4]int{0, 1, 134, 0},
			ok:         true,
		},
		{
			name:       "product four part",
			version:    "0.1.134.2",
			normalized: "0.1.134.2",
			parts:      [4]int{0, 1, 134, 2},
			ok:         true,
		},
		{
			name:       "local release tag",
			version:    "v0.1.134-local.1",
			normalized: "0.1.134.1",
			parts:      [4]int{0, 1, 134, 1},
			ok:         true,
		},
		{name: "empty", version: "", ok: false},
		{name: "two part", version: "0.1", ok: false},
		{name: "five part", version: "0.1.134.1.1", ok: false},
		{name: "non numeric", version: "0.1.x", ok: false},
		{name: "missing local revision", version: "v0.1.134-local", ok: false},
		{name: "non numeric local revision", version: "v0.1.134-local.x", ok: false},
		{name: "leading zero local revision", version: "v0.1.134-local.01", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, parts, ok := parseProductVersion(tt.version)
			require.Equal(t, tt.ok, ok)
			if !tt.ok {
				return
			}
			require.Equal(t, tt.normalized, normalized)
			require.Equal(t, tt.parts, parts)
		})
	}
}

func TestUpdateServiceFetchLatestReleaseNormalizesLocalTagAndUsesProjectRepo(t *testing.T) {
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "v0.1.134-local.1",
			Name:    "v0.1.134-local.1",
		},
	}
	svc := NewUpdateService(&updateServiceCacheStub{}, client, "0.1.133.1", "release")

	info, err := svc.CheckUpdate(context.Background(), true)

	require.NoError(t, err)
	require.Equal(t, githubRepo, client.repo)
	require.Equal(t, "consider908/sub2api", client.repo)
	require.Equal(t, "0.1.133.1", info.CurrentVersion)
	require.Equal(t, "0.1.134.1", info.LatestVersion)
	require.True(t, info.HasUpdate)
}

func TestUpdateServiceCheckUpdateWarnsOnInvalidLatestReleaseTag(t *testing.T) {
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "v0.1.134-local.x",
			Name:    "v0.1.134-local.x",
		},
	}
	svc := NewUpdateService(&updateServiceCacheStub{}, client, "0.1.133.1", "release")

	info, err := svc.CheckUpdate(context.Background(), true)

	require.NoError(t, err)
	require.Equal(t, "0.1.133.1", info.CurrentVersion)
	require.Equal(t, "0.1.133.1", info.LatestVersion)
	require.False(t, info.HasUpdate)
	require.Contains(t, info.Warning, "invalid latest release tag")
}

func TestUpdateServiceCheckUpdateUsesCachedDataWhenLatestReleaseInvalid(t *testing.T) {
	cache := &updateServiceCacheStub{
		data: `{"latest":"0.1.134.1","release_info":{"name":"cached"},"timestamp":4102444800}`,
	}
	client := &updateServiceGitHubClientStub{
		release: &GitHubRelease{
			TagName: "invalid",
			Name:    "invalid",
		},
	}
	svc := NewUpdateService(cache, client, "0.1.133.1", "release")

	info, err := svc.CheckUpdate(context.Background(), true)

	require.NoError(t, err)
	require.Equal(t, "0.1.134.1", info.LatestVersion)
	require.True(t, info.HasUpdate)
	require.True(t, info.Cached)
	require.Contains(t, info.Warning, "Using cached data")
}
