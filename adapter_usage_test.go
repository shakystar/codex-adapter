package codexadapter

import (
	"encoding/json"
	"testing"
)

func TestDecodeLastUsage(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Usage
	}{
		{name: "missing tokenUsage"},
		{name: "null tokenUsage", raw: `null`},
		{name: "missing last", raw: `{}`},
		{name: "null last", raw: `{"last":null}`},
		{name: "empty last", raw: `{"last":{}}`},
		{name: "missing input", raw: `{"last":{"outputTokens":0,"cachedInputTokens":0}}`},
		{name: "missing output", raw: `{"last":{"inputTokens":0,"cachedInputTokens":0}}`},
		{name: "missing cached", raw: `{"last":{"inputTokens":0,"outputTokens":0}}`},
		{name: "null input", raw: `{"last":{"inputTokens":null,"outputTokens":0,"cachedInputTokens":0}}`},
		{name: "null output", raw: `{"last":{"inputTokens":0,"outputTokens":null,"cachedInputTokens":0}}`},
		{name: "null cached", raw: `{"last":{"inputTokens":0,"outputTokens":0,"cachedInputTokens":null}}`},
		{name: "negative input", raw: `{"last":{"inputTokens":-1,"outputTokens":0,"cachedInputTokens":0}}`},
		{name: "negative output", raw: `{"last":{"inputTokens":0,"outputTokens":-1,"cachedInputTokens":0}}`},
		{name: "negative cached", raw: `{"last":{"inputTokens":0,"outputTokens":0,"cachedInputTokens":-1}}`},
		{name: "string input", raw: `{"last":{"inputTokens":"7","outputTokens":0,"cachedInputTokens":0}}`},
		{name: "fractional output", raw: `{"last":{"inputTokens":0,"outputTokens":1.5,"cachedInputTokens":0}}`},
		{name: "overflowing cached", raw: `{"last":{"inputTokens":0,"outputTokens":0,"cachedInputTokens":9223372036854775808}}`},
		{name: "non-object last", raw: `{"last":[]}`},
		{name: "non-object tokenUsage", raw: `[]`},
		{
			name: "explicit zeros",
			raw:  `{"last":{"inputTokens":0,"outputTokens":0,"cachedInputTokens":0}}`,
			want: Usage{Known: true},
		},
		{
			name: "measured counters",
			raw:  `{"last":{"inputTokens":7,"outputTokens":3,"cachedInputTokens":2}}`,
			want: Usage{Known: true, InputTokens: 7, OutputTokens: 3, CachedInputTokens: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeLastUsage(json.RawMessage(tt.raw)); got != tt.want {
				t.Errorf("decodeLastUsage(%q) = %+v, want %+v", tt.raw, got, tt.want)
			}
		})
	}
}
