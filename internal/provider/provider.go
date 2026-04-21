package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dynu/terraform-provider-dynu/internal/dynuclient"
)

var _ provider.Provider = &dynuProvider{}

type dynuProvider struct {
	version string
}

type dynuProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

type providerData struct {
	client *dynuclient.Client
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dynuProvider{version: version}
	}
}

func (p *dynuProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dynu"
	resp.Version = p.version
}

func (p *dynuProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Dynu DNS read-only operations.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Dynu API key. Can also be provided using DYNU_API_KEY.",
			},
		},
	}
}

func (p *dynuProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data dynuProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("DYNU_API_KEY")
	if !data.APIKey.IsNull() {
		apiKey = data.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Dynu API key",
			"Set api_key in the provider configuration or DYNU_API_KEY in the environment.",
		)
		return
	}

	providerData := &providerData{client: dynuclient.New(apiKey)}
	resp.DataSourceData = providerData
	resp.ResourceData = nil
}

func (p *dynuProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainsDataSource,
		NewDomainDataSource,
		NewDNSRecordsDataSource,
	}
}

func (p *dynuProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
