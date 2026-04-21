package provider

import (
	"errors"

	"github.com/dynu/terraform-provider-dynu/internal/dynuclient"
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
		return defaultSummary + " (not found)"
	default:
		return defaultSummary
	}
}
