// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_controltower_controls")
func DataSourceControls() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: DataSourceControlsRead,

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

func DataSourceControlsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	targetIdentifier := d.Get("target_identifier").(string)

	var controls []string
	var nextToken string

	for {
		input := &controltower.ListEnabledControlsInput{
			TargetIdentifier: aws.String(targetIdentifier),
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		out, err := conn.ListEnabledControls(ctx, input)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, control := range out.EnabledControls {
			controls = append(controls, aws.ToString(control.ControlIdentifier))
		}

		if out.NextToken == nil {
			break
		}

		nextToken = aws.ToString(out.NextToken)
	}

	d.SetId(targetIdentifier)
	d.Set("enabled_controls", controls)

	return nil
}
