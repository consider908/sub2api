package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/httpclient"
	kiropkg "github.com/Wei-Shaw/sub2api/internal/pkg/kiro"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"go.uber.org/zap"
)

type kiroAvailableProfile struct {
	ARN         string `json:"arn"`
	ProfileName string `json:"profileName"`
}

type kiroListAvailableProfilesResponse struct {
	Profiles  []kiroAvailableProfile `json:"profiles"`
	NextToken string                 `json:"nextToken"`
}

func (r *kiroListAvailableProfilesResponse) firstARN() string {
	for _, profile := range r.Profiles {
		if arn := strings.TrimSpace(profile.ARN); arn != "" {
			return arn
		}
	}
	return ""
}

var kiroProfileResolutionFlight singleflight.Group
var kiroHTTPClientFactory = httpclient.GetClient

func kiroListAvailableProfiles(ctx context.Context, account *Account, token string) (*kiroListAvailableProfilesResponse, error) {
	if account == nil {
		return nil, fmt.Errorf("account is nil")
	}
	region := kiroAPIRegion(account)
	endpointURL := fmt.Sprintf("https://q.%s.amazonaws.com/", region)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, strings.NewReader(`{"maxResults":10}`))
	if err != nil {
		return nil, fmt.Errorf("create list profiles request: %w", err)
	}
	accountKey := buildKiroAccountKey(account)
	machineID := buildKiroMachineID(account)
	req.Header.Set("Content-Type", "application/x-amz-json-1.0")
	req.Header.Set("X-Amz-Target", "AmazonCodeWhispererService.ListAvailableProfiles")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", kiropkg.BuildRuntimeUserAgent(accountKey, machineID))
	req.Header.Set("X-Amz-User-Agent", kiropkg.BuildRuntimeAmzUserAgent(accountKey, machineID))
	req.Header.Set("Amz-Sdk-Request", "attempt=1; max=1")
	req.Header.Set("Amz-Sdk-Invocation-Id", uuid.NewString())
	applyKiroConditionalHeaders(req, account)

	client, err := kiroHTTPClientFactory(httpclient.Options{
		ProxyURL:           kiroProxyURL(account),
		Timeout:            30 * time.Second,
		ValidateResolvedIP: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create http client: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list available profiles request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list available profiles: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed kiroListAvailableProfilesResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &parsed, nil
}

func kiroResolveAndPersistProfileArn(ctx context.Context, repo AccountRepository, account *Account, token string) string {
	if account == nil || account.Type != AccountTypeOAuth {
		return ""
	}
	if isKiroRelayAccount(account) || isKiroDirectApiKeyAccount(account) {
		return ""
	}

	existingARN := strings.TrimSpace(account.GetCredential("profile_arn"))
	if existingARN != "" && !kiroIsPlaceholderProfileARN(existingARN) {
		return existingARN
	}

	accountID := account.ID
	result, err, _ := kiroProfileResolutionFlight.Do(fmt.Sprintf("%d", accountID), func() (any, error) {
		profiles, err := kiroListAvailableProfiles(ctx, account, token)
		if err != nil {
			logger.L().Warn("kiro profileArn resolution failed",
				zap.Int64("account_id", accountID),
				zap.Error(err),
			)
			return "", err
		}

		arn := profiles.firstARN()
		if arn == "" {
			arn = kiroDefaultProfileARN(account)
		}
		if account.Credentials == nil {
			account.Credentials = map[string]any{}
		}
		account.Credentials["profile_arn"] = arn
		if repo != nil {
			if updater, ok := any(repo).(accountCredentialsUpdater); ok {
				account.Credentials = cloneCredentials(account.Credentials)
				if err := updater.UpdateCredentials(ctx, account.ID, account.Credentials); err != nil {
					logger.L().Warn("kiro profileArn persist failed",
						zap.Int64("account_id", accountID),
						zap.Error(err),
					)
				}
			}
		}
		return arn, nil
	})
	if err != nil {
		return existingARN
	}

	if resolvedARN, ok := result.(string); ok && strings.TrimSpace(resolvedARN) != "" {
		return resolvedARN
	}
	return existingARN
}

func (s *GatewayService) resolveAndPersistKiroProfileArn(ctx context.Context, account *Account, token string) string {
	return kiroResolveAndPersistProfileArn(ctx, s.accountRepo, account, token)
}

func (s *GatewayService) ensureKiroProfileArnForRequest(ctx context.Context, account *Account, token, mode string) {
	if account == nil || mode != KiroEndpointModeKRS || account.Type != AccountTypeOAuth {
		return
	}
	existingARN := strings.TrimSpace(account.GetCredential("profile_arn"))
	if existingARN != "" && !kiroIsPlaceholderProfileARN(existingARN) {
		return
	}
	_ = s.resolveAndPersistKiroProfileArn(ctx, account, token)
}
