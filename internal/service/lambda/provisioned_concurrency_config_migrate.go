// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceProvisionedConcurrencyConfigV0() *schema.Resource {
	// Resource with v0 schema (provider v5.3.0 and below)
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"provisioned_concurrent_executions": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"qualifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func provisionedConcurrencyConfigStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		rawState = map[string]interface{}{}
	}

	// Convert id separator from ":" to ","
	parts := []string{
		rawState["function_name"].(string),
		rawState["qualifier"].(string),
	}

	id, err := flex.FlattenResourceId(parts, provisionedConcurrencyConfigResourceIDPartCount, false)
	if err != nil {
		return rawState, err
	}
	rawState[names.AttrID] = id

	return rawState, nil
}
