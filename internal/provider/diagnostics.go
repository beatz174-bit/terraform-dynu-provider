package provider

import (
	"errors"
	"strings"

	"github.com/beatz174-bit/terraform-provider-dynu/internal/dynuclient"
)

func diagnosticSummary(defaultSummary string, err error) string {
	var apiErr *dynuclient.APIError
	if !errors.As(err, &apiErr) {
		return defaultSummary
	}

	switch apiErr.StatusCode {
	case 401, 403:
		return defaultSummary + " (authentication failed)"
	case 404:
		normalizedType := strings.ToLower(strings.TrimSpace(apiErr.Type))
		if strings.Contains(normalizedType, "not found") {
			return defaultSummary + " (not found)"
		}
		return defaultSummary
	default:
		return defaultSummary
	}
}
