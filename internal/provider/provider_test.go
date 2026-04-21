package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		config types.String
		env    string
		want   string
	}{
		{name: "config wins", config: types.StringValue("config-key"), env: "env-key", want: "config-key"},
		{name: "env fallback", config: types.StringNull(), env: "env-key", want: "env-key"},
		{name: "trim spaces", config: types.StringValue("  config-key "), env: " env-key ", want: "config-key"},
		{name: "unknown uses env", config: types.StringUnknown(), env: "env-key", want: "env-key"},
		{name: "empty when missing", config: types.StringNull(), env: "", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveAPIKey(tc.config, tc.env); got != tc.want {
				t.Fatalf("resolveAPIKey() = %q, want %q", got, tc.want)
			}
		})
	}
}
