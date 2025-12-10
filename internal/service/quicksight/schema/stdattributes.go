// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"regexp"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func AWSAccountIDAttribute() fwschema.StringAttribute { // nosemgrep:ci.aws-in-func-name
	return fwschema.StringAttribute{
		Optional: true,
		Computed: true,
		Validators: []validator.String{
			fwvalidators.AWSAccountID(),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func AWSAccountIDSchema() *sdkschema.Schema { // nosemgrep:ci.aws-in-func-name
	return &sdkschema.Schema{
		Type:         sdkschema.TypeString,
		Optional:     true,
		Computed:     true,
		ForceNew:     true,
		ValidateFunc: verify.ValidAccountID,
	}
}

func AWSAccountIDDataSourceSchema() *sdkschema.Schema { // nosemgrep:ci.aws-in-func-name
	return &sdkschema.Schema{
		Type:         sdkschema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: verify.ValidAccountID,
	}
}

const (
	DefaultNamespace = "default"
)

func NamespaceAttribute() fwschema.StringAttribute {
	return fwschema.StringAttribute{
		Optional: true,
		Computed: true,
		Default:  stringdefault.StaticString(DefaultNamespace),
		Validators: []validator.String{
			stringvalidator.LengthBetween(1, 64),
			stringvalidator.RegexMatches(namespaceRegex()),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func NamespaceSchema() *sdkschema.Schema {
	return &sdkschema.Schema{
		Type:     sdkschema.TypeString,
		Optional: true,
		ForceNew: true,
		Default:  DefaultNamespace,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 64),
			validation.StringMatch(namespaceRegex()),
		),
	}
}

func NamespaceDataSourceSchema() *sdkschema.Schema {
	return &sdkschema.Schema{
		Type:     sdkschema.TypeString,
		Optional: true,
		Default:  DefaultNamespace,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 64),
			validation.StringMatch(namespaceRegex()),
		),
	}
}

func namespaceRegex() (*regexp.Regexp, string) {
	return regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "must contain only alphanumeric characters, hyphens, underscores, and periods"
}
