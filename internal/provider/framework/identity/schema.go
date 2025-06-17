// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identity

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func NewIdentitySchema(identitySpec inttypes.Identity) identityschema.Schema {
	schemaAttrs := make(map[string]identityschema.Attribute, len(identitySpec.Attributes))
	for _, attr := range identitySpec.Attributes {
		schemaAttrs[attr.Name] = newIdentityAttribute(attr)
	}
	return identityschema.Schema{
		Attributes: schemaAttrs,
	}
}

func newIdentityAttribute(attribute inttypes.IdentityAttribute) identityschema.Attribute {
	attr := identityschema.StringAttribute{}
	if attribute.Required {
		attr.RequiredForImport = true
	} else {
		attr.OptionalForImport = true
	}
	return attr
}
