// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_connect_prompt")
func DataSourcePrompt() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePromptRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
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

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get("instance_id").(string)
	name := d.Get("name").(string)

	promptSummary, err := dataSourceGetPromptSummaryByName(ctx, conn, instanceID, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "finding Connect Prompt Summary by name (%s): %s", name, err)
	}

	if promptSummary == nil {
		return sdkdiag.AppendErrorf(diags, "finding Connect Prompt Summary by name (%s): not found", name)
	}

	d.Set("arn", promptSummary.Arn)
	d.Set("instance_id", instanceID)
	d.Set("prompt_id", promptSummary.Id)
	d.Set("name", promptSummary.Name)

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.ToString(promptSummary.Id)))

	return diags
}

func dataSourceGetPromptSummaryByName(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.PromptSummary, error) {
	var result *awstypes.PromptSummary

	input := &connect.ListPromptsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(ListPromptsMaxResults),
	}

	pages := connect.NewListPromptsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, cf := range page.PromptSummaryList {
			if aws.ToString(cf.Name) == name {
				result = &cf
			}
		}
	}

	return result, nil
}
