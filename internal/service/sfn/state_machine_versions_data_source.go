// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_sfn_state_machine_versions", name="State Machine Versions")
func dataSourceStateMachineVersions() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStateMachineVersionsRead,

		Schema: map[string]*schema.Schema{
			"statemachine_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"statemachine_versions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceStateMachineVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	smARN := d.Get("statemachine_arn").(string)
	input := &sfn.ListStateMachineVersionsInput{
		StateMachineArn: aws.String(smARN),
	}
	var smvARNs []string

	err := listStateMachineVersionsPages(ctx, conn, input, func(page *sfn.ListStateMachineVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.StateMachineVersions {
			smvARNs = append(smvARNs, aws.ToString(v.StateMachineVersionArn))
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Step Functions State Machine (%s) Versions: %s", smARN, err)
	}

	d.SetId(smARN)
	d.Set("statemachine_versions", smvARNs)

	return diags
}
