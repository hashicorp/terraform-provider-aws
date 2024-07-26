// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sfn_state_machine", name="State Machine")
func dataSourceStateMachine() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStateMachineRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationDate: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := findStateByARN(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Step Functions State Machines: %s", err)
	}

	if n := len(output); n == 0 {
		return sdkdiag.AppendErrorf(diags, "no Step Functions State Machines matched")
	} else if n > 1 {
		return sdkdiag.AppendErrorf(diags, "%d Step Functions State Machines matched; use additional constraints to reduce matches to a single State Machine", n)
	}

	out, err := findStateMachineByARN(ctx, conn, aws.ToString(output[0].StateMachineArn))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Step Functions State Machine (%s): %s", aws.ToString(output[0].StateMachineArn), err)
	}

	d.SetId(aws.ToString(out.StateMachineArn))
	d.Set(names.AttrARN, out.StateMachineArn)
	d.Set(names.AttrCreationDate, out.CreationDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, out.Description)
	d.Set("definition", out.Definition)
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrRoleARN, out.RoleArn)
	d.Set("revision_id", out.RevisionId)
	d.Set(names.AttrStatus, out.Status)

	return diags
}

func findStateByARN(ctx context.Context, conn *sfn.Client, name string) ([]awstypes.StateMachineListItem, error) {
	var output []awstypes.StateMachineListItem

	pages := sfn.NewListStateMachinesPaginator(conn, &sfn.ListStateMachinesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.StateMachineDoesNotExist](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: name,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.StateMachines {
			if name == aws.ToString(v.Name) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
