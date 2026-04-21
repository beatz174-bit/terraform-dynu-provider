package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &domainsDataSource{}
	_ datasource.DataSourceWithConfigure = &domainsDataSource{}
)

type domainsDataSource struct {
	clientProvider *providerData
}

type domainsDataSourceModel struct {
	Domains []domainModel `tfsdk:"domains"`
}

type domainModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	UnicodeName       types.String `tfsdk:"unicode_name"`
	Token             types.String `tfsdk:"token"`
	State             types.String `tfsdk:"state"`
	Group             types.String `tfsdk:"group"`
	IPv4Address       types.String `tfsdk:"ipv4_address"`
	IPv6Address       types.String `tfsdk:"ipv6_address"`
	TTL               types.Int64  `tfsdk:"ttl"`
	IPv4              types.Bool   `tfsdk:"ipv4"`
	IPv6              types.Bool   `tfsdk:"ipv6"`
	IPv4WildcardAlias types.Bool   `tfsdk:"ipv4_wildcard_alias"`
	IPv6WildcardAlias types.Bool   `tfsdk:"ipv6_wildcard_alias"`
	AllowZoneTransfer types.Bool   `tfsdk:"allow_zone_transfer"`
	DNSSEC            types.Bool   `tfsdk:"dnssec"`
	CreatedOn         types.String `tfsdk:"created_on"`
	UpdatedOn         types.String `tfsdk:"updated_on"`
}

func NewDomainsDataSource() datasource.DataSource {
	return &domainsDataSource{}
}

func (d *domainsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

func (d *domainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List DNS domains from Dynu.",
		Attributes: map[string]schema.Attribute{
			"domains": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "List of Dynu DNS domains.",
				NestedObject: schema.NestedAttributeObject{Attributes: domainAttributes()},
			},
		},
	}
}

func (d *domainsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source configure type", fmt.Sprintf("Expected *providerData, got %T", req.ProviderData))
		return
	}

	d.clientProvider = providerData
}

func (d *domainsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	domains, err := d.clientProvider.client.ListDomains(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Dynu domains", err.Error())
		return
	}

	state := domainsDataSourceModel{Domains: make([]domainModel, 0, len(domains))}
	for _, domain := range domains {
		state.Domains = append(state.Domains, mapDomain(domain))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func domainAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":                  schema.Int64Attribute{Computed: true},
		"name":                schema.StringAttribute{Computed: true},
		"unicode_name":        schema.StringAttribute{Computed: true},
		"token":               schema.StringAttribute{Computed: true, Sensitive: true},
		"state":               schema.StringAttribute{Computed: true},
		"group":               schema.StringAttribute{Computed: true},
		"ipv4_address":        schema.StringAttribute{Computed: true},
		"ipv6_address":        schema.StringAttribute{Computed: true},
		"ttl":                 schema.Int64Attribute{Computed: true},
		"ipv4":                schema.BoolAttribute{Computed: true},
		"ipv6":                schema.BoolAttribute{Computed: true},
		"ipv4_wildcard_alias": schema.BoolAttribute{Computed: true},
		"ipv6_wildcard_alias": schema.BoolAttribute{Computed: true},
		"allow_zone_transfer": schema.BoolAttribute{Computed: true},
		"dnssec":              schema.BoolAttribute{Computed: true},
		"created_on":          schema.StringAttribute{Computed: true},
		"updated_on":          schema.StringAttribute{Computed: true},
	}
}
