// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sfn_alias", name="Alias")
func dataSourceAlias() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAliasRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
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
						names.AttrWeight: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	aliasArn := ""
	in := &sfn.ListStateMachineAliasesInput{
		StateMachineArn: aws.String(d.Get("statemachine_arn").(string)),
	}

	out, err := conn.ListStateMachineAliases(ctx, in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Step Functions State Machines: %s", err)
	}

	if n := len(out.StateMachineAliases); n == 0 {
		return sdkdiag.AppendErrorf(diags, "no Step Functions State Machine Aliases matched")
	}

	for _, in := range out.StateMachineAliases {
		if v := aws.ToString(in.StateMachineAliasArn); strings.HasSuffix(v, d.Get(names.AttrName).(string)) {
			aliasArn = v
		}
	}

	if aliasArn == "" {
		return sdkdiag.AppendErrorf(diags, "no Step Functions State Machine Aliases matched")
	}

	output, err := findAliasByARN(ctx, conn, aliasArn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Step Functions State Machine Alias (%s): %s", aliasArn, err)
	}

	d.SetId(aliasArn)
	d.Set(names.AttrARN, output.StateMachineAliasArn)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrCreationDate, aws.ToTime(output.CreationDate).Format(time.RFC3339))

	if err := d.Set("routing_configuration", flattenAliasRoutingConfiguration(output.RoutingConfiguration)); err != nil {
		return create.AppendDiagError(diags, names.SFN, create.ErrActionSetting, ResNameAlias, d.Id(), err)
	}

	return diags
}
