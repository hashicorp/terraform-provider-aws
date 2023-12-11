// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/controltower/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_controltower_controls", name="Control")
func dataSourceControls() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceControlsRead,

		Schema: map[string]*schema.Schema{
			"enabled_controls": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"target_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func dataSourceControlsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	targetIdentifier := d.Get("target_identifier").(string)
	input := &controltower.ListEnabledControlsInput{
		TargetIdentifier: aws.String(targetIdentifier),
	}

	controls, err := findEnabledControls(ctx, conn, input, tfslices.PredicateTrue[*types.EnabledControlSummary]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ControlTower Controls (%s): %s", targetIdentifier, err)
	}

	d.SetId(targetIdentifier)
	d.Set("enabled_controls", tfslices.ApplyToAll(controls, func(v *types.EnabledControlSummary) string {
		return aws.ToString(v.ControlIdentifier)
	}))

	return diags
}
