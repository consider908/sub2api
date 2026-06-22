package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGatewayServiceRecordUsage_PersistsKiroCredits(t *testing.T) {
	svc := &GatewayService{}
	apiKey := &APIKey{ID: 501}
	user := &User{ID: 601}
	account := &Account{ID: 701, Platform: PlatformKiro, Type: AccountTypeOAuth}
	result := &ForwardResult{
		RequestID: "kiro_credits_usage_log",
		Usage: ClaudeUsage{
			InputTokens:  10,
			OutputTokens: 6,
			KiroCredits:  0.17,
		},
		Model:    "kiro-agent",
		Duration: time.Second,
	}

	log := svc.buildRecordUsageLog(
		context.Background(),
		&recordUsageCoreInput{},
		result,
		apiKey,
		user,
		account,
		nil,
		result.Model,
		1,
		1,
		1,
		BillingTypeBalance,
		false,
		nil,
		&recordUsageOpts{},
	)

	require.NotNil(t, log)
	require.NotNil(t, log.KiroCredits)
	require.InDelta(t, 0.17, *log.KiroCredits, 1e-12)
	require.Equal(t, 10, log.InputTokens)
	require.Equal(t, 6, log.OutputTokens)
	require.Equal(t, 0.0, log.TotalCost)
	require.Equal(t, 0.0, log.ActualCost)
}
