// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sfn_alias")
func DataSourceAlias() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAliasRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"routing_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"state_machine_version_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"statemachine_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	DSNameAlias = "Alias Data Source"
)

func dataSourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)
	aliasArn := ""

	in := &sfn.ListStateMachineAliasesInput{
		StateMachineArn: aws.String(d.Get("statemachine_arn").(string)),
	}

	out, err := conn.ListStateMachineAliasesWithContext(ctx, in)

	if err != nil {
		return diag.Errorf("listing Step Functions State Machines: %s", err)
	}

	if n := len(out.StateMachineAliases); n == 0 {
		return diag.Errorf("no Step Functions State Machine Aliases matched")
	}

	for _, in := range out.StateMachineAliases {
		if v := aws.StringValue(in.StateMachineAliasArn); strings.HasSuffix(v, d.Get("name").(string)) {
			aliasArn = v
		}
	}

	if aliasArn == "" {
		return diag.Errorf("no Step Functions State Machine Aliases matched")
	}

	output, err := FindAliasByARN(ctx, conn, aliasArn)

	if err != nil {
		return diag.Errorf("reading Step Functions State Machine Alias (%s): %s", aliasArn, err)
	}

	d.SetId(aliasArn)
	d.Set("arn", output.StateMachineAliasArn)
	d.Set("name", output.Name)
	d.Set("description", output.Description)
	d.Set("creation_date", aws.TimeValue(output.CreationDate).Format(time.RFC3339))

	if err := d.Set("routing_configuration", flattenAliasRoutingConfiguration(output.RoutingConfiguration)); err != nil {
		return create.DiagError(names.SFN, create.ErrActionSetting, ResNameAlias, d.Id(), err)
	}

	return nil
}
