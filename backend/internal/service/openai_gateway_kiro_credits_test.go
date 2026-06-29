package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractOpenAIUsageFromJSONBytes_ParsesKiroCreditsAliases(t *testing.T) {
	tests := []struct {
		name string
		body string
		want float64
	}{
		{
			name: "internal alias",
			body: `{"response":{"usage":{"input_tokens":1,"output_tokens":2,"_sub2api_kiro_credits":1.25}}}`,
			want: 1.25,
		},
		{
			name: "camel alias",
			body: `{"usage":{"input_tokens":1,"output_tokens":2,"kiroCredits":2.5}}`,
			want: 2.5,
		},
		{
			name: "generic alias",
			body: `{"message":{"usage":{"input_tokens":1,"output_tokens":2,"consumedCredits":3.75}}}`,
			want: 3.75,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			usage, ok := extractOpenAIUsageFromJSONBytes([]byte(tc.body))
			require.True(t, ok)
			require.InDelta(t, tc.want, usage.KiroCredits, 1e-12)
		})
	}
}

func TestStripInternalKiroCreditsJSONBytes_RemovesOnlyInternalField(t *testing.T) {
	body := []byte(`{"response":{"usage":{"input_tokens":1,"output_tokens":2,"_sub2api_kiro_credits":1.25,"kiro_credits":1.25}}}`)
	sanitized := stripInternalKiroCreditsJSONBytes(body)

	require.NotContains(t, string(sanitized), "_sub2api_kiro_credits")
	require.Contains(t, string(sanitized), `"kiro_credits":1.25`)
}
