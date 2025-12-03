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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	DSNameStateMachines = "State Machines Data Source"
)

// @SDKDataSource("aws_sfn_state_machines")
func DataSourceStateMachines() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStateMachinesV2,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceStateMachinesV2(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	input := &sfn.ListStateMachinesInput{}

	var arns []string
	var state_machine_names []string

	err := conn.ListStateMachinesPagesWithContext(ctx, input, func(page *sfn.ListStateMachinesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, stateMachine := range page.StateMachines {
			if stateMachine == nil {
				continue
			}

			arns = append(arns, aws.StringValue(stateMachine.StateMachineArn))
			state_machine_names = append(state_machine_names, aws.StringValue(stateMachine.Name))
		}

		return !lastPage
	})
	if err != nil {
		return create.DiagError(names.SFN, create.ErrActionReading, DSNameStateMachines, "", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("arns", arns)
	d.Set("names", state_machine_names)

	return nil
}
