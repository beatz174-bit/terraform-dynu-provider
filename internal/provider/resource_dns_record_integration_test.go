package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/dynu/terraform-provider-dynu/internal/testutil/fakedynu"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIntegrationResourceDNSRecordLifecycleAndImport(t *testing.T) {
	ctx := context.Background()
	fake := fakedynu.NewServer()
	defer fake.Close()

	r := NewDNSRecordResource().(*dnsRecordResource)
	configureResource(t, r, fake.BaseURL())

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	createPlan := dnsRecordResourceModel{
		Hostname:   types.StringValue("api.a.example.com"),
		RecordType: types.StringValue("TXT"),
		Content:    types.StringValue("created"),
		TTL:        types.Int64Value(60),
		State:      types.BoolValue(true),
		Group:      types.StringValue("test"),
		Host:       types.StringNull(),
		NodeName:   types.StringNull(),
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(ctx, &createPlan); diags.HasError() {
		t.Fatalf("set plan diagnostics: %v", diags)
	}

	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create diagnostics: %v", createResp.Diagnostics)
	}

	var state dnsRecordResourceModel
	if diags := createResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("state get diagnostics: %v", diags)
	}
	if !strings.HasPrefix(state.ID.ValueString(), "1001/") {
		t.Fatalf("expected composite id, got %q", state.ID.ValueString())
	}
	if state.Content.ValueString() != "created" {
		t.Fatalf("unexpected created content: %q", state.Content.ValueString())
	}

	updatePlan := state
	updatePlan.Content = types.StringValue("updated")
	updatePlan.TTL = types.Int64Value(600)
	plan = tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(ctx, &updatePlan); diags.HasError() {
		t.Fatalf("set update plan diagnostics: %v", diags)
	}

	updateResp := resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: plan}, &updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %v", updateResp.Diagnostics)
	}
	if diags := updateResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("updated state get diagnostics: %v", diags)
	}
	if state.Content.ValueString() != "updated" || state.TTL.ValueInt64() != 600 {
		t.Fatalf("unexpected updated state: %#v", state)
	}

	importResp := resource.ImportStateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.ImportState(ctx, resource.ImportStateRequest{ID: state.ID.ValueString()}, &importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("import diagnostics: %v", importResp.Diagnostics)
	}
	var importedID string
	if diags := importResp.State.GetAttribute(ctx, path.Root("id"), &importedID); diags.HasError() {
		t.Fatalf("imported id diagnostics: %v", diags)
	}
	if importedID != state.ID.ValueString() {
		t.Fatalf("expected imported id %q, got %q", state.ID.ValueString(), importedID)
	}

	deleteResp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: updateResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("delete diagnostics: %v", deleteResp.Diagnostics)
	}
}

func TestIntegrationResourceDNSRecordReadNotFoundRemovesState(t *testing.T) {
	ctx := context.Background()
	fake := fakedynu.NewServer()
	defer fake.Close()

	r := NewDNSRecordResource().(*dnsRecordResource)
	configureResource(t, r, fake.BaseURL())

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	state := dnsRecordResourceModel{
		ID: types.StringValue("1001/10"),
	}
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	if diags := tfState.Set(ctx, &state); diags.HasError() {
		t.Fatalf("set state diagnostics: %v", diags)
	}

	fake.DeleteRecord(1001, 10)

	readResp := resource.ReadResponse{State: tfState}
	r.Read(ctx, resource.ReadRequest{State: tfState}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
	}

	var after dnsRecordResourceModel
	diags := readResp.State.Get(ctx, &after)
	if !diags.HasError() {
		t.Fatal("expected diagnostics because resource was removed from state")
	}
}

func TestIntegrationResourceDNSRecordImportStateInvalidID(t *testing.T) {
	ctx := context.Background()
	r := NewDNSRecordResource().(*dnsRecordResource)
	importResp := resource.ImportStateResponse{}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "bad-id"}, &importResp)
	if !importResp.Diagnostics.HasError() {
		t.Fatal("expected import diagnostics for malformed ID")
	}
	if !strings.Contains(importResp.Diagnostics[0].Detail(), "domain_id/record_id") {
		t.Fatalf("unexpected diagnostic detail: %s", importResp.Diagnostics[0].Detail())
	}
}

func TestIntegrationResourceDNSRecordDynamicAStateStableAndTransitionToStatic(t *testing.T) {
	ctx := context.Background()
	fake := fakedynu.NewServer()
	defer fake.Close()

	r := NewDNSRecordResource().(*dnsRecordResource)
	configureResource(t, r, fake.BaseURL())

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	createPlan := dnsRecordResourceModel{
		Hostname:   types.StringValue("api.a.example.com"),
		RecordType: types.StringValue("A"),
		Content:    types.StringNull(),
		TTL:        types.Int64Value(60),
		State:      types.BoolValue(true),
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(ctx, &createPlan); diags.HasError() {
		t.Fatalf("set plan diagnostics: %v", diags)
	}

	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create diagnostics: %v", createResp.Diagnostics)
	}

	var state dnsRecordResourceModel
	if diags := createResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("state get diagnostics: %v", diags)
	}
	if !state.Dynamic.ValueBool() {
		t.Fatalf("expected dynamic=true for blank A, got %#v", state.Dynamic)
	}
	if !state.Content.IsNull() {
		t.Fatalf("expected dynamic A content to remain null in state, got %q", state.Content.ValueString())
	}

	readResp := resource.ReadResponse{State: createResp.State}
	r.Read(ctx, resource.ReadRequest{State: createResp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", readResp.Diagnostics)
	}
	if diags := readResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("read state diagnostics: %v", diags)
	}
	if !state.Content.IsNull() {
		t.Fatalf("expected dynamic A read to preserve null content, got %q", state.Content.ValueString())
	}

	updatePlan := state
	updatePlan.Content = types.StringValue("192.0.2.42")
	updatePlan.Dynamic = types.BoolValue(false)

	plan = tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(ctx, &updatePlan); diags.HasError() {
		t.Fatalf("set update plan diagnostics: %v", diags)
	}

	updateResp := resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: plan}, &updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %v", updateResp.Diagnostics)
	}
	if diags := updateResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("updated state diagnostics: %v", diags)
	}
	if state.Dynamic.ValueBool() {
		t.Fatalf("expected static A after setting content")
	}
	if state.Content.ValueString() != "192.0.2.42" {
		t.Fatalf("expected static content after update, got %q", state.Content.ValueString())
	}
}

func configureResource(t *testing.T, r resource.ResourceWithConfigure, baseURL string) {
	t.Helper()
	resp := resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: &providerData{client: newDynuClient("dummy-local-key", types.StringValue(baseURL))},
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("configure diagnostics: %v", resp.Diagnostics)
	}
}
