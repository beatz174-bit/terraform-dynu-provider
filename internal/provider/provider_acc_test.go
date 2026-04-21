package provider

import (
	"os"
	"testing"
)

func TestAccScaffold(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" || os.Getenv("DYNU_API_KEY") == "" {
		t.Skip("set TF_ACC=1 and DYNU_API_KEY to enable acceptance tests")
	}

	// Optional: DYNU_DOMAIN may be used by future acceptance test cases.
	_ = os.Getenv("DYNU_DOMAIN")

	// Acceptance tests for read-only data sources are intentionally scaffolded in phase 1.
	// Add terraform-plugin-testing based test cases in a follow-up with stable fixtures.
}
