package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func hostnameFromConfig(ctx context.Context, config tfsdk.Config) (types.String, error) {
	var hostname types.String
	diags := config.GetAttribute(ctx, path.Root("hostname"), &hostname)
	if diags.HasError() {
		return types.StringNull(), fmt.Errorf("decode hostname from data source config: %s", diags.Errors()[0].Summary())
	}

	return hostname, nil
}
