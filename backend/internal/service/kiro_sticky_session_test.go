//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func kiroStickyGroup(enabled bool) *Group {
	return &Group{Platform: PlatformKiro, KiroAutoStickyEnabled: enabled}
}

func nonKiroStickyGroup() *Group {
	return &Group{Platform: PlatformAnthropic}
}

func kiroStickyBody(systemPrompt string, turns int) []byte {
	msgs := `[{"role":"user","content":"hello"}`
	for i := 1; i < turns; i++ {
		msgs += `,{"role":"assistant","content":"reply"},{"role":"user","content":"next turn"}`
	}
	msgs += `]`
	return []byte(`{"model":"claude-sonnet-4-5","system":"` + systemPrompt + `","messages":` + msgs + `}`)
}

func kiroStickyBodyWithSession(field, value string) []byte {
	return []byte(`{"model":"claude-sonnet-4-5","` + field + `":"` + value + `","system":"stable","messages":[{"role":"user","content":"hello"}]}`)
}

func parseKiroStickyRequest(t *testing.T, body []byte, group *Group, apiKeyID int64) *ParsedRequest {
	t.Helper()
	parsed, err := ParseGatewayRequest(NewRequestBodyRef(body), "anthropic")
	require.NoError(t, err)
	parsed.Group = group
	parsed.SessionContext = &SessionContext{APIKeyID: apiKeyID}
	return parsed
}

func TestKiroStickySession_AutoSystemPromptStableAcrossTurns(t *testing.T) {
	svc := &GatewayService{}

	hash1 := svc.GenerateSessionHash(parseKiroStickyRequest(t, kiroStickyBody("stable system", 1), kiroStickyGroup(true), 42))
	hash2 := svc.GenerateSessionHash(parseKiroStickyRequest(t, kiroStickyBody("stable system", 3), kiroStickyGroup(true), 42))

	require.NotEmpty(t, hash1)
	require.Equal(t, hash1, hash2)
}

func TestKiroStickySession_DisabledSkipsAutoInference(t *testing.T) {
	svc := &GatewayService{}
	parsed := parseKiroStickyRequest(t, kiroStickyBody("stable system", 1), kiroStickyGroup(false), 42)

	require.Empty(t, svc.GenerateSessionHash(parsed))
}

func TestKiroStickySession_ExplicitSessionIDWorksWhenAutoDisabled(t *testing.T) {
	svc := &GatewayService{}
	parsedA := parseKiroStickyRequest(t, kiroStickyBody("system a", 1), kiroStickyGroup(false), 42)
	parsedB := parseKiroStickyRequest(t, kiroStickyBody("system b", 1), kiroStickyGroup(false), 42)
	parsedA.ExplicitSessionID = "manual-session"
	parsedB.ExplicitSessionID = "manual-session"

	require.Equal(t, svc.GenerateSessionHash(parsedA), svc.GenerateSessionHash(parsedB))
}

func TestKiroStickySession_BodySessionIDTakesPrecedence(t *testing.T) {
	svc := &GatewayService{}
	hashA := svc.GenerateSessionHash(parseKiroStickyRequest(t, kiroStickyBodyWithSession("conversation_id", "conv-1"), kiroStickyGroup(true), 42))
	hashB := svc.GenerateSessionHash(parseKiroStickyRequest(t, kiroStickyBodyWithSession("conversation_id", "conv-1"), kiroStickyGroup(false), 42))
	hashC := svc.GenerateSessionHash(parseKiroStickyRequest(t, kiroStickyBodyWithSession("conversation_id", "conv-2"), kiroStickyGroup(true), 42))

	require.NotEmpty(t, hashA)
	require.Equal(t, hashA, hashB)
	require.NotEqual(t, hashA, hashC)
}

func TestKiroStickySession_DifferentAPIKeysAreIsolated(t *testing.T) {
	svc := &GatewayService{}
	body := kiroStickyBodyWithSession("session_id", "shared-session")

	require.NotEqual(t,
		svc.GenerateSessionHash(parseKiroStickyRequest(t, body, kiroStickyGroup(true), 1)),
		svc.GenerateSessionHash(parseKiroStickyRequest(t, body, kiroStickyGroup(true), 2)),
	)
}

func TestKiroStickySession_NonKiroKeepsMessageHashBehavior(t *testing.T) {
	svc := &GatewayService{}
	hash1 := svc.GenerateSessionHash(parseKiroStickyRequest(t, kiroStickyBody("stable system", 1), nonKiroStickyGroup(), 42))
	hash2 := svc.GenerateSessionHash(parseKiroStickyRequest(t, kiroStickyBody("stable system", 3), nonKiroStickyGroup(), 42))

	require.NotEmpty(t, hash1)
	require.NotEqual(t, hash1, hash2)
}

func TestExtractBodySessionID(t *testing.T) {
	cases := []struct {
		name string
		body []byte
		want string
	}{
		{name: "prompt cache key", body: []byte(`{"prompt_cache_key":"pcache-1","conversation_id":"conv-1"}`), want: "pcache-1"},
		{name: "camel prompt cache key", body: []byte(`{"promptCacheKey":"pcache-camel","conversation_id":"conv-1"}`), want: "pcache-camel"},
		{name: "conversation id", body: []byte(`{"conversation_id":"conv-1"}`), want: "conv-1"},
		{name: "camel session id", body: []byte(`{"sessionId":"sess-1"}`), want: "sess-1"},
		{name: "metadata prompt cache key", body: []byte(`{"metadata":{"prompt_cache_key":"meta-pcache-1","threadId":"thread-meta-1"}}`), want: "meta-pcache-1"},
		{name: "metadata thread id", body: []byte(`{"metadata":{"threadId":"thread-meta-1"}}`), want: "thread-meta-1"},
		{name: "metadata session id", body: []byte(`{"metadata":{"session_id":"meta-sess-1"}}`), want: "meta-sess-1"},
		{name: "blank ignored", body: []byte(`{"session_id":"   ","thread_id":"thread-1"}`), want: "thread-1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, extractBodySessionID(tc.body))
		})
	}
}

func TestKiroStickySessionTTLNormalization(t *testing.T) {
	cases := []struct {
		name string
		in   *Group
		want int
	}{
		{name: "default", in: &Group{Platform: PlatformKiro}, want: DefaultKiroStickySessionTTLSeconds},
		{name: "min", in: &Group{Platform: PlatformKiro, KiroStickySessionTTLSeconds: 1}, want: MinKiroStickySessionTTLSeconds},
		{name: "max", in: &Group{Platform: PlatformKiro, KiroStickySessionTTLSeconds: 999999}, want: MaxKiroStickySessionTTLSeconds},
		{name: "custom", in: &Group{Platform: PlatformKiro, KiroStickySessionTTLSeconds: 7200}, want: 7200},
		{name: "non kiro", in: &Group{Platform: PlatformAnthropic, KiroStickySessionTTLSeconds: 7200}, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, tc.in.EffectiveKiroStickySessionTTLSeconds())
		})
	}
}
