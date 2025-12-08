// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceUserInGroupV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrGroupName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserGroupName,
			},
			names.AttrUserPoolID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserPoolID,
			},
			names.AttrUsername: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func userInGroupStateUpgradeV0(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	rawState[names.AttrID] = fmt.Sprintf("%s,%s,%s", rawState[names.AttrUserPoolID], rawState[names.AttrGroupName], rawState[names.AttrUsername])

	return rawState, nil
}
