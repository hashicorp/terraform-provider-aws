// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package identity

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func NewIdentitySchema(identitySpec inttypes.Identity) map[string]*schema.Schema {
	identitySchema := make(map[string]*schema.Schema, len(identitySpec.Attributes))
	for _, attr := range identitySpec.Attributes {
		identitySchema[attr.Name()] = newIdentityAttribute(attr)
	}
	return identitySchema
}

func newIdentityAttribute(attribute inttypes.IdentityAttribute) *schema.Schema {
	attr := &schema.Schema{
		Type: schema.TypeString,
	}
	switch attribute.IdentityType() {
	case inttypes.BoolIdentityType:
		attr.Type = schema.TypeBool
	case inttypes.FloatIdentityType:
		attr.Type = schema.TypeFloat
	case inttypes.IntIdentityType:
		attr.Type = schema.TypeInt
	}

	if attribute.Required() {
		attr.RequiredForImport = true
	} else {
		attr.OptionalForImport = true
	}
	return attr
}
