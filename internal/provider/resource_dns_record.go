package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dynu/terraform-provider-dynu/internal/dynuclient"
)

var (
	_ resource.Resource                   = &dnsRecordResource{}
	_ resource.ResourceWithConfigure      = &dnsRecordResource{}
	_ resource.ResourceWithImportState    = &dnsRecordResource{}
	_ resource.ResourceWithValidateConfig = &dnsRecordResource{}
)

type dnsRecordResource struct {
	clientProvider *providerData
}

type dnsRecordResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Hostname   types.String `tfsdk:"hostname"`
	RecordType types.String `tfsdk:"record_type"`
	Content    types.String `tfsdk:"content"`
	TTL        types.Int64  `tfsdk:"ttl"`
	State      types.Bool   `tfsdk:"state"`
	Group      types.String `tfsdk:"group"`
	Host       types.String `tfsdk:"host"`
	NodeName   types.String `tfsdk:"node_name"`
	DomainID   types.Int64  `tfsdk:"domain_id"`
	DomainName types.String `tfsdk:"domain_name"`
	UpdatedOn  types.String `tfsdk:"updated_on"`
}

func NewDNSRecordResource() resource.Resource {
	return &dnsRecordResource{}
}

func (r *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dynu DNS record for the root domain resolved from hostname.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, Description: "Resource identifier in domain_id/record_id format."},
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "Fully-qualified hostname under the target root domain.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(hostnameValidator, "must be a valid fully-qualified hostname"),
				},
			},
			"record_type": schema.StringAttribute{Required: true, Description: "DNS record type (A, AAAA, CNAME, TXT, etc.).", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
			"content":     schema.StringAttribute{Optional: true, Description: "Record content/value."},
			"ttl": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "DNS TTL in seconds.",
				Validators:  []validator.Int64{int64validator.AtLeast(0)},
			},
			"state":       schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether this DNS record is active."},
			"group":       schema.StringAttribute{Optional: true, Computed: true, Description: "Dynu group value for this record."},
			"host":        schema.StringAttribute{Optional: true, Computed: true, Description: "Host field for supported Dynu record types."},
			"node_name":   schema.StringAttribute{Optional: true, Computed: true, Description: "Node/label portion of the record."},
			"domain_id":   schema.Int64Attribute{Computed: true, Description: "Dynu domain ID resolved from hostname."},
			"domain_name": schema.StringAttribute{Computed: true, Description: "Dynu root domain name resolved from hostname."},
			"updated_on":  schema.StringAttribute{Computed: true, Description: "Last update timestamp as returned by Dynu."},
		},
	}
}

func (r *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configure type", fmt.Sprintf("Expected *providerData, got %T", req.ProviderData))
		return
	}
	r.clientProvider = providerData
}

func (r *dnsRecordResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config dnsRecordResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	recordType, skip := knownNormalizedString(config.RecordType)
	if skip {
		return
	}
	if recordType == "A" || recordType == "AAAA" {
		return
	}
	if config.Content.IsUnknown() {
		return
	}
	if config.Content.IsNull() || strings.TrimSpace(config.Content.ValueString()) == "" {
		resp.Diagnostics.AddError(
			"Missing required content for DNS record type",
			fmt.Sprintf("The %q record type requires a non-empty content value. Set the content attribute or use A/AAAA when content should be omitted.", recordType),
		)
	}
}

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, domainName, err := r.clientProvider.client.GetRootDomain(ctx, plan.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(diagnosticSummary("Unable to resolve Dynu domain from hostname", err), err.Error())
		return
	}

	createReq := dynuclient.CreateDNSRecordRequest{
		NodeName:   recordNodeName(plan.NodeName, plan.Hostname, domainName),
		RecordType: strings.TrimSpace(plan.RecordType.ValueString()),
		Content:    stringPointerFromOptional(plan.Content),
		TTL:        int64FromOptional(plan.TTL),
		State:      boolPointerFromOptional(plan.State),
		Group:      stringFromOptional(plan.Group),
		Host:       stringFromOptional(plan.Host),
	}
	if !validateDNSRecordContentForType(createReq.RecordType, createReq.Content, &resp.Diagnostics) {
		return
	}

	record, err := r.clientProvider.client.CreateDNSRecord(ctx, domainID, createReq)
	if err != nil {
		addDNSRecordWriteDiagnostic("create", createReq.RecordType, createReq.Content, err, &resp.Diagnostics)
		return
	}

	state := mapDNSRecordToState(*record)
	if plan.Content.IsNull() || plan.Content.IsUnknown() {
		state.Content = types.StringNull()
	}
	state.ID = types.StringValue(formatDNSRecordID(record.DomainID, record.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, recordID, err := parseDNSRecordID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid resource ID", err.Error())
		return
	}

	record, err := r.clientProvider.client.GetDNSRecord(ctx, domainID, recordID)
	if err != nil {
		var apiErr *dynuclient.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(diagnosticSummary("Unable to read Dynu DNS record", err), err.Error())
		return
	}

	nextState := mapDNSRecordToState(*record)
	if state.Content.IsNull() {
		nextState.Content = types.StringNull()
	}
	nextState.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &nextState)...)
}

func (r *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, recordID, err := parseDNSRecordID(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid resource ID", err.Error())
		return
	}

	domainName := strings.TrimSpace(plan.DomainName.ValueString())
	if domainName == "" {
		_, resolvedDomainName, err := r.clientProvider.client.GetRootDomain(ctx, plan.Hostname.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(diagnosticSummary("Unable to resolve Dynu domain from hostname", err), err.Error())
			return
		}
		domainName = resolvedDomainName
	}

	updateReq := dynuclient.UpdateDNSRecordRequest{
		NodeName:   recordNodeName(plan.NodeName, plan.Hostname, domainName),
		RecordType: strings.TrimSpace(plan.RecordType.ValueString()),
		Content:    stringPointerFromOptional(plan.Content),
		TTL:        int64FromOptional(plan.TTL),
		State:      boolPointerFromOptional(plan.State),
		Group:      stringFromOptional(plan.Group),
		Host:       stringFromOptional(plan.Host),
	}
	if !validateDNSRecordContentForType(updateReq.RecordType, updateReq.Content, &resp.Diagnostics) {
		return
	}

	if _, err := r.clientProvider.client.UpdateDNSRecord(ctx, domainID, recordID, updateReq); err != nil {
		addDNSRecordWriteDiagnostic("update", updateReq.RecordType, updateReq.Content, err, &resp.Diagnostics)
		return
	}

	record, err := r.clientProvider.client.GetDNSRecord(ctx, domainID, recordID)
	if err != nil {
		resp.Diagnostics.AddError(diagnosticSummary("Unable to read Dynu DNS record", err), err.Error())
		return
	}

	nextState := mapDNSRecordToState(*record)
	if plan.Content.IsNull() || plan.Content.IsUnknown() {
		nextState.Content = types.StringNull()
	}
	nextState.ID = types.StringValue(formatDNSRecordID(record.DomainID, record.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &nextState)...)
}

func (r *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, recordID, err := parseDNSRecordID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid resource ID", err.Error())
		return
	}

	err = r.clientProvider.client.DeleteDNSRecord(ctx, domainID, recordID)
	if err == nil {
		return
	}
	var apiErr *dynuclient.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
		return
	}
	resp.Diagnostics.AddError(diagnosticSummary("Unable to delete Dynu DNS record", err), err.Error())
}

func (r *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	domainID, recordID, err := parseDNSRecordID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	state := dnsRecordResourceModel{ID: types.StringValue(formatDNSRecordID(domainID, recordID))}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapDNSRecordToState(record dynuclient.DNSRecord) dnsRecordResourceModel {
	return dnsRecordResourceModel{
		Hostname:   mapString(record.Hostname),
		RecordType: mapString(record.RecordType),
		Content:    mapString(record.Content),
		TTL:        types.Int64Value(record.TTL),
		State:      types.BoolValue(record.State),
		Group:      mapString(record.Group),
		Host:       mapString(record.Host),
		NodeName:   mapString(record.NodeName),
		DomainID:   types.Int64Value(record.DomainID),
		DomainName: mapString(record.DomainName),
		UpdatedOn:  mapString(record.UpdatedOn),
	}
}

func parseDNSRecordID(id string) (int64, int64, error) {
	parts := strings.Split(strings.TrimSpace(id), "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected import ID in domain_id/record_id format")
	}
	domainID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || domainID <= 0 {
		return 0, 0, fmt.Errorf("invalid domain_id in ID %q", id)
	}
	recordID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || recordID <= 0 {
		return 0, 0, fmt.Errorf("invalid record_id in ID %q", id)
	}
	return domainID, recordID, nil
}

func formatDNSRecordID(domainID int64, recordID int64) string {
	return fmt.Sprintf("%d/%d", domainID, recordID)
}

func recordNodeName(nodeName types.String, hostname types.String, domainName string) string {
	if !nodeName.IsNull() && !nodeName.IsUnknown() {
		return strings.TrimSpace(nodeName.ValueString())
	}
	host := strings.TrimSpace(hostname.ValueString())
	domain := strings.TrimSpace(domainName)
	if strings.EqualFold(host, domain) {
		return ""
	}
	suffix := "." + domain
	if strings.HasSuffix(strings.ToLower(host), strings.ToLower(suffix)) {
		return strings.TrimSuffix(host, suffix)
	}
	return host
}

func int64FromOptional(value types.Int64) int64 {
	if value.IsNull() || value.IsUnknown() {
		return 0
	}
	return value.ValueInt64()
}

func boolPointerFromOptional(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueBool()
	return &v
}

func stringFromOptional(value types.String) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return strings.TrimSpace(value.ValueString())
}

func stringPointerFromOptional(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	trimmed := strings.TrimSpace(value.ValueString())
	return &trimmed
}

func validateDNSRecordContentForType(recordType string, content *string, diagnostics *diag.Diagnostics) bool {
	normalizedType := strings.ToUpper(strings.TrimSpace(recordType))
	if normalizedType == "A" || normalizedType == "AAAA" {
		return true
	}

	if content == nil || strings.TrimSpace(*content) == "" {
		diagnostics.AddError(
			"Missing required content for DNS record type",
			fmt.Sprintf("The %q record type requires a non-empty content value. Set the content attribute or choose a type that supports omitted content (A/AAAA).", normalizedType),
		)
		return false
	}

	return true
}

func addDNSRecordWriteDiagnostic(operation string, recordType string, content *string, err error, diagnostics *diag.Diagnostics) {
	detail := err.Error()
	var apiErr *dynuclient.APIError
	if errors.As(err, &apiErr) {
		presence := "omitted"
		if content != nil {
			presence = fmt.Sprintf("set to %q", *content)
		}
		detail = fmt.Sprintf("%s. Dynu rejected this %s request for record type %q where content was %s.", err.Error(), operation, strings.ToUpper(strings.TrimSpace(recordType)), presence)
	}

	diagnostics.AddError(diagnosticSummary(fmt.Sprintf("Unable to %s Dynu DNS record", operation), err), detail)
}

func knownNormalizedString(value types.String) (string, bool) {
	if value.IsNull() || value.IsUnknown() {
		return "", true
	}
	return strings.ToUpper(strings.TrimSpace(value.ValueString())), false
}
