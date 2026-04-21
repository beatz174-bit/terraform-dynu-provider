package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &domainDataSource{}
	_ datasource.DataSourceWithConfigure = &domainDataSource{}
)

type domainDataSource struct {
	clientProvider *providerData
}

type domainDataSourceModel struct {
	Hostname types.String `tfsdk:"hostname"`
	Domain   types.Object `tfsdk:"domain"`
}

func NewDomainDataSource() datasource.DataSource {
	return &domainDataSource{}
}

func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *domainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a Dynu DNS domain by hostname.",
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "Hostname to resolve to a root domain.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"domain": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Resolved Dynu DNS domain details.",
				Attributes:  domainAttributes(),
			},
		},
	}
}

func (d *domainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state domainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Hostname.IsUnknown() || state.Hostname.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("hostname"), "Invalid hostname", "hostname must be known and non-null.")
		return
	}

	domainID, _, err := d.clientProvider.client.GetRootDomain(ctx, state.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to resolve Dynu domain from hostname", err.Error())
		return
	}

	domain, err := d.clientProvider.client.GetDomainByID(ctx, domainID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Dynu domain", err.Error())
		return
	}

	domainValue := mapDomain(*domain)
	domainObject, diags := types.ObjectValue(
		map[string]attr.Type{
			"id":                  types.Int64Type,
			"name":                types.StringType,
			"unicode_name":        types.StringType,
			"token":               types.StringType,
			"state":               types.StringType,
			"group":               types.StringType,
			"ipv4_address":        types.StringType,
			"ipv6_address":        types.StringType,
			"ttl":                 types.Int64Type,
			"ipv4":                types.BoolType,
			"ipv6":                types.BoolType,
			"ipv4_wildcard_alias": types.BoolType,
			"ipv6_wildcard_alias": types.BoolType,
			"allow_zone_transfer": types.BoolType,
			"dnssec":              types.BoolType,
			"created_on":          types.StringType,
			"updated_on":          types.StringType,
		},
		map[string]attr.Value{
			"id":                  domainValue.ID,
			"name":                domainValue.Name,
			"unicode_name":        domainValue.UnicodeName,
			"token":               domainValue.Token,
			"state":               domainValue.State,
			"group":               domainValue.Group,
			"ipv4_address":        domainValue.IPv4Address,
			"ipv6_address":        domainValue.IPv6Address,
			"ttl":                 domainValue.TTL,
			"ipv4":                domainValue.IPv4,
			"ipv6":                domainValue.IPv6,
			"ipv4_wildcard_alias": domainValue.IPv4WildcardAlias,
			"ipv6_wildcard_alias": domainValue.IPv6WildcardAlias,
			"allow_zone_transfer": domainValue.AllowZoneTransfer,
			"dnssec":              domainValue.DNSSEC,
			"created_on":          domainValue.CreatedOn,
			"updated_on":          domainValue.UpdatedOn,
		},
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Domain = domainObject

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
