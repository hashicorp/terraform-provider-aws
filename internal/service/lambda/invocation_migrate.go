// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceInvocationConfigV0() *schema.Resource {
	// Resource with v0 schema (provider v5.83.0 and below)
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"input": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"lifecycle_scope": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          lifecycleScopeCreateOnly,
				ValidateDiagFunc: enum.Validate[lifecycleScope](),
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  FunctionVersionLatest,
			},
			"result": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"terraform_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "tf",
			},
			names.AttrTriggers: {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// This upgrader _should have_ been implemented in v5.1.0 alongside the
// additions of the lifecycle_scope and terraform_key arguments with
// default values, but was overlooked at the time. Because we cannot
// reasonably go back and patch the schemas for all versions in between,
// we instead handle both pre-v5.1.0 and v5.1.0-v5.83.0 versions of the
// previous state.
func invocationStateUpgradeV0(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	// If upgrading from a version < v5.1.0, these values will be unset and
	// should be set to the defaults added in v5.1.0
	if rawState["lifecycle_scope"] == nil {
		rawState["lifecycle_scope"] = lifecycleScopeCreateOnly
	}
	if rawState["terraform_key"] == nil {
		rawState["terraform_key"] = "tf"
	}

	return rawState, nil
}
