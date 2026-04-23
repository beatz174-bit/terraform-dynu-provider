package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func hostnameFromConfig(config tfsdk.Config) (types.String, error) {
	if !config.Raw.IsKnown() {
		return types.StringUnknown(), nil
	}
	if config.Raw.IsNull() {
		return types.StringNull(), nil
	}

	attributes := map[string]tftypes.Value{}
	if err := config.Raw.As(&attributes); err != nil {
		return types.StringNull(), fmt.Errorf("decode data source config object: %w", err)
	}

	hostnameValue, ok := attributes["hostname"]
	if !ok {
		return types.StringNull(), fmt.Errorf("hostname is missing from data source config")
	}
	if !hostnameValue.IsKnown() {
		return types.StringUnknown(), nil
	}
	if hostnameValue.IsNull() {
		return types.StringNull(), nil
	}

	var hostname string
	if err := hostnameValue.As(&hostname); err != nil {
		return types.StringNull(), fmt.Errorf("decode hostname: %w", err)
	}

	return types.StringValue(hostname), nil
}
