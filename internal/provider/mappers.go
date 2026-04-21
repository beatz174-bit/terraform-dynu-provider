package provider

import (
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dynu/terraform-provider-dynu/internal/dynuclient"
)

var domainObjectTypes = map[string]attr.Type{
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
}

func mapDomain(domain dynuclient.Domain) domainModel {
	return domainModel{
		ID:                types.Int64Value(domain.ID),
		Name:              types.StringValue(domain.Name),
		UnicodeName:       types.StringValue(domain.UnicodeName),
		Token:             types.StringValue(domain.Token),
		State:             types.StringValue(domain.State),
		Group:             types.StringValue(domain.Group),
		IPv4Address:       mapString(domain.IPv4Address),
		IPv6Address:       mapString(domain.IPv6Address),
		TTL:               types.Int64Value(domain.TTL),
		IPv4:              types.BoolValue(domain.IPv4),
		IPv6:              types.BoolValue(domain.IPv6),
		IPv4WildcardAlias: types.BoolValue(domain.IPv4WildcardAlias),
		IPv6WildcardAlias: types.BoolValue(domain.IPv6WildcardAlias),
		AllowZoneTransfer: types.BoolValue(domain.AllowZoneTransfer),
		DNSSEC:            types.BoolValue(domain.DNSSEC),
		CreatedOn:         mapString(domain.CreatedOn),
		UpdatedOn:         mapString(domain.UpdatedOn),
	}
}

func domainObjectValue(domain dynuclient.Domain) (types.Object, diag.Diagnostics) {
	mapped := mapDomain(domain)

	return types.ObjectValue(domainObjectTypes, map[string]attr.Value{
		"id":                  mapped.ID,
		"name":                mapped.Name,
		"unicode_name":        mapped.UnicodeName,
		"token":               mapped.Token,
		"state":               mapped.State,
		"group":               mapped.Group,
		"ipv4_address":        mapped.IPv4Address,
		"ipv6_address":        mapped.IPv6Address,
		"ttl":                 mapped.TTL,
		"ipv4":                mapped.IPv4,
		"ipv6":                mapped.IPv6,
		"ipv4_wildcard_alias": mapped.IPv4WildcardAlias,
		"ipv6_wildcard_alias": mapped.IPv6WildcardAlias,
		"allow_zone_transfer": mapped.AllowZoneTransfer,
		"dnssec":              mapped.DNSSEC,
		"created_on":          mapped.CreatedOn,
		"updated_on":          mapped.UpdatedOn,
	})
}

func mapString(in string) types.String {
	if in == "" {
		return types.StringNull()
	}
	return types.StringValue(in)
}

func sortDomains(domains []dynuclient.Domain) {
	sort.Slice(domains, func(i, j int) bool {
		if domains[i].Name != domains[j].Name {
			return domains[i].Name < domains[j].Name
		}
		return domains[i].ID < domains[j].ID
	})
}

func sortDNSRecords(records []dynuclient.DNSRecord) {
	sort.Slice(records, func(i, j int) bool {
		if records[i].Hostname != records[j].Hostname {
			return records[i].Hostname < records[j].Hostname
		}
		if records[i].RecordType != records[j].RecordType {
			return records[i].RecordType < records[j].RecordType
		}
		return records[i].ID < records[j].ID
	})
}
