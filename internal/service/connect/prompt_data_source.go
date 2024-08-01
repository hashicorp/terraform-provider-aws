// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_prompt")
func DataSourcePrompt() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePromptRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"prompt_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePromptRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)

	promptSummary, err := dataSourceGetPromptSummaryByName(ctx, conn, instanceID, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "finding Connect Prompt Summary by name (%s): %s", name, err)
	}

	if promptSummary == nil {
		return sdkdiag.AppendErrorf(diags, "finding Connect Prompt Summary by name (%s): not found", name)
	}

	d.Set(names.AttrARN, promptSummary.Arn)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set("prompt_id", promptSummary.Id)
	d.Set(names.AttrName, promptSummary.Name)

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(promptSummary.Id)))

	return diags
}

func dataSourceGetPromptSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.PromptSummary, error) {
	var result *connect.PromptSummary

	input := &connect.ListPromptsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListPromptsMaxResults),
	}

	err := conn.ListPromptsPagesWithContext(ctx, input, func(page *connect.ListPromptsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.PromptSummaryList {
			if cf == nil {
				continue
			}

			if aws.StringValue(cf.Name) == name {
				result = cf
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
