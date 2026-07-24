// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package identity

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func NewIdentitySchema(identitySpec inttypes.Identity) identityschema.Schema {
	schemaAttrs := make(map[string]identityschema.Attribute, len(identitySpec.Attributes))
	for _, attr := range identitySpec.Attributes {
		schemaAttrs[attr.Name()] = newIdentityAttribute(attr)
	}
	return identityschema.Schema{
		Attributes: schemaAttrs,
	}
}

func newIdentityAttribute(attribute inttypes.IdentityAttribute) identityschema.Attribute {
	required := attribute.Required()
	var optional bool
	if !required {
		optional = true
	}

	identityAttributes := map[inttypes.IdentityType]identityschema.Attribute{
		inttypes.BoolIdentityType: identityschema.BoolAttribute{
			RequiredForImport: required,
			OptionalForImport: optional,
		},
		inttypes.FloatIdentityType: identityschema.Float32Attribute{
			RequiredForImport: required,
			OptionalForImport: optional,
		},
		inttypes.Float64IdentityType: identityschema.Float64Attribute{
			RequiredForImport: required,
			OptionalForImport: optional,
		},
		inttypes.IntIdentityType: identityschema.Int32Attribute{
			RequiredForImport: required,
			OptionalForImport: optional,
		},
		inttypes.Int64IdentityType: identityschema.Int64Attribute{
			RequiredForImport: required,
			OptionalForImport: optional,
		},
		inttypes.StringIdentityType: identityschema.StringAttribute{
			RequiredForImport: required,
			OptionalForImport: optional,
		},
	}

	return identityAttributes[attribute.IdentityType()]
}
