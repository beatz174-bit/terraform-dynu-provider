package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestHostnameFromConfigIgnoresComputedAttributesDomain(t *testing.T) {
	ds := NewDomainDataSource().(*domainDataSource)
	var schemaResp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)

	configType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"hostname": tftypes.String,
		"domain": tftypes.DynamicPseudoType,
	}}
	configValue := tftypes.NewValue(configType, map[string]tftypes.Value{
		"hostname": tftypes.NewValue(tftypes.String, "www.example.com"),
		"domain":   tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
	})

	hostname, err := hostnameFromConfig(context.Background(), tfsdk.Config{Schema: schemaResp.Schema, Raw: configValue})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostname.ValueString() != "www.example.com" {
		t.Fatalf("unexpected hostname: %q", hostname.ValueString())
	}
}

func TestHostnameFromConfigIgnoresComputedAttributesDNSRecords(t *testing.T) {
	ds := NewDNSRecordsDataSource().(*dnsRecordsDataSource)
	var schemaResp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)

	configType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"hostname":    tftypes.String,
		"domain_id":   tftypes.Number,
		"domain_name": tftypes.String,
		"records":     tftypes.DynamicPseudoType,
	}}
	configValue := tftypes.NewValue(configType, map[string]tftypes.Value{
		"hostname":    tftypes.NewValue(tftypes.String, "www.example.com"),
		"domain_id":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"domain_name": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"records":     tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
	})

	hostname, err := hostnameFromConfig(context.Background(), tfsdk.Config{Schema: schemaResp.Schema, Raw: configValue})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostname.ValueString() != "www.example.com" {
		t.Fatalf("unexpected hostname: %q", hostname.ValueString())
	}
}
