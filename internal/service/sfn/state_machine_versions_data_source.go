// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_sfn_state_machine_versions")
func DataSourceStateMachineVersions() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStateMachineVersionsRead,

		Schema: map[string]*schema.Schema{
			"statemachine_versions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"statemachine_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	DSNameStateMachineVersions = "StateMachine versions Data Source"
)

func dataSourceStateMachineVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	in := &sfn.ListStateMachineVersionsInput{
		StateMachineArn: aws.String(d.Get("statemachine_arn").(string)),
	}

	var statemachine_version_arns []string

	for {
		output, err := conn.ListStateMachineVersionsWithContext(ctx, in)
		if err != nil {
			return diag.Errorf("listing Step Functions State Machine versions: %s", err)
		}
		for _, in := range output.StateMachineVersions {
			statemachine_version_arns = append(statemachine_version_arns, *in.StateMachineVersionArn)
		}
		if output.NextToken == nil {
			break
		}
		in.NextToken = output.NextToken
	}

	if n := len(statemachine_version_arns); n == 0 {
		return diag.Errorf("no Step Functions State Machine versions matched")
	}

	d.SetId(d.Get("statemachine_arn").(string))
	d.Set("statemachine_versions", statemachine_version_arns)

	return nil
}
