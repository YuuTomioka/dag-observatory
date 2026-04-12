package di

import "testing"

func TestResolveSwaggerUIEnabled(t *testing.T) {
	tests := []struct {
		name string
		env  string
		raw  string
		want bool
	}{
		{
			name: "dev default enabled",
			env:  "dev",
			raw:  "",
			want: true,
		},
		{
			name: "staging default disabled",
			env:  "staging",
			raw:  "",
			want: false,
		},
		{
			name: "prod default disabled",
			env:  "prod",
			raw:  "",
			want: false,
		},
		{
			name: "explicit true overrides env default",
			env:  "prod",
			raw:  "true",
			want: true,
		},
		{
			name: "explicit false overrides env default",
			env:  "dev",
			raw:  "false",
			want: false,
		},
		{
			name: "invalid value falls back to env default",
			env:  "dev",
			raw:  "invalid",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveSwaggerUIEnabled(tt.env, tt.raw)
			if got != tt.want {
				t.Fatalf("resolveSwaggerUIEnabled(%q, %q) = %v, want %v", tt.env, tt.raw, got, tt.want)
			}
		})
	}
}
