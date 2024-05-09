// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sfn_state_machine")
func DataSourceStateMachine() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStateMachineRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"definition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceStateMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	name := d.Get(names.AttrName).(string)
	var arns []string

	err := conn.ListStateMachinesPagesWithContext(ctx, &sfn.ListStateMachinesInput{}, func(page *sfn.ListStateMachinesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.StateMachines {
			if aws.StringValue(v.Name) == name {
				arns = append(arns, aws.StringValue(v.StateMachineArn))
			}
		}

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("listing Step Functions State Machines: %s", err)
	}

	if n := len(arns); n == 0 {
		return diag.Errorf("no Step Functions State Machines matched")
	} else if n > 1 {
		return diag.Errorf("%d Step Functions State Machines matched; use additional constraints to reduce matches to a single State Machine", n)
	}

	arn := arns[0]
	output, err := FindStateMachineByARN(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("reading Step Functions State Machine (%s): %s", arn, err)
	}

	d.SetId(arn)
	d.Set(names.AttrARN, output.StateMachineArn)
	d.Set("creation_date", output.CreationDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set("definition", output.Definition)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrRoleARN, output.RoleArn)
	d.Set("revision_id", output.RevisionId)
	d.Set(names.AttrStatus, output.Status)

	return nil
}
