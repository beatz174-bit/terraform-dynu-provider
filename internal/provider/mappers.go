package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dynu/terraform-provider-dynu/internal/dynuclient"
)

func mapDomain(domain dynuclient.Domain) domainModel {
	return domainModel{
		ID:                types.Int64Value(domain.ID),
		Name:              types.StringValue(domain.Name),
		UnicodeName:       types.StringValue(domain.UnicodeName),
		Token:             types.StringValue(domain.Token),
		State:             types.StringValue(domain.State),
		Group:             types.StringValue(domain.Group),
		IPv4Address:       types.StringValue(domain.IPv4Address),
		IPv6Address:       types.StringValue(domain.IPv6Address),
		TTL:               types.Int64Value(domain.TTL),
		IPv4:              types.BoolValue(domain.IPv4),
		IPv6:              types.BoolValue(domain.IPv6),
		IPv4WildcardAlias: types.BoolValue(domain.IPv4WildcardAlias),
		IPv6WildcardAlias: types.BoolValue(domain.IPv6WildcardAlias),
		AllowZoneTransfer: types.BoolValue(domain.AllowZoneTransfer),
		DNSSEC:            types.BoolValue(domain.DNSSEC),
		CreatedOn:         types.StringValue(domain.CreatedOn),
		UpdatedOn:         types.StringValue(domain.UpdatedOn),
	}
}

func mapString(in string) types.String {
	if in == "" {
		return types.StringNull()
	}
	return types.StringValue(in)
}
