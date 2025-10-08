// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_prompt", name="Prompt")
func dataSourcePrompt() *schema.Resource {
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

func dataSourcePromptRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)
	promptSummary, err := findPromptSummaryByTwoPartKey(ctx, conn, instanceID, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Prompt (%s) summary: %s", name, err)
	}

	promptID := aws.ToString(promptSummary.Id)
	id := promptCreateResourceID(instanceID, promptID)
	d.SetId(id)
	d.Set(names.AttrARN, promptSummary.Arn)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, promptSummary.Name)
	d.Set("prompt_id", promptID)

	return diags
}

const promptResourceIDSeparator = ":"

func promptCreateResourceID(instanceID, promptID string) string {
	parts := []string{instanceID, promptID}
	id := strings.Join(parts, promptResourceIDSeparator)

	return id
}

func findPromptSummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.PromptSummary, error) {
	const maxResults = 60
	input := &connect.ListPromptsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findPromptSummary(ctx, conn, input, func(v *awstypes.PromptSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findPromptSummary(ctx context.Context, conn *connect.Client, input *connect.ListPromptsInput, filter tfslices.Predicate[*awstypes.PromptSummary]) (*awstypes.PromptSummary, error) {
	output, err := findPromptSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPromptSummaries(ctx context.Context, conn *connect.Client, input *connect.ListPromptsInput, filter tfslices.Predicate[*awstypes.PromptSummary]) ([]awstypes.PromptSummary, error) {
	var output []awstypes.PromptSummary

	pages := connect.NewListPromptsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.PromptSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
