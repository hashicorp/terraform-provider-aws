// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_securityhub_standards_control")
func newStandardsControlResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceStandardsControl(), framework.SDKListSpec[awstypes.StandardsControl, listStandardsControlModel]{
		ConfigSchema: listStandardsControlConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listStandardsControlModel) (iter.Seq2[awstypes.StandardsControl, error], diag.Diagnostics) {
			input := securityhub.DescribeStandardsControlsInput{
				StandardsSubscriptionArn: fwflex.StringFromFramework(ctx, query.StandardsSubscriptionARN),
			}

			return listStandardsControls(ctx, c.SecurityHubClient(ctx), &input), nil
		},
		ID: func(sc awstypes.StandardsControl, _ listStandardsControlModel) string {
			return aws.ToString(sc.StandardsControlArn)
		},
		Flatten:     resourceStandardsControlFlatten,
		DisplayName: func(sc awstypes.StandardsControl) string { return aws.ToString(sc.Title) },
	})
}

type listStandardsControlModel struct {
	framework.WithRegionModel
	StandardsSubscriptionARN fwtypes.ARN `tfsdk:"standards_subscription_arn"`
}

func listStandardsControls(ctx context.Context, conn *securityhub.Client, input *securityhub.DescribeStandardsControlsInput) iter.Seq2[awstypes.StandardsControl, error] {
	return func(yield func(awstypes.StandardsControl, error) bool) {
		pages := securityhub.NewDescribeStandardsControlsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.StandardsControl](), fmt.Errorf("listing Security Hub Standards Controls: %w", err))
				return
			}

			for _, item := range page.Controls {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

func listStandardsControlConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"standards_subscription_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN that represents your subscription to a supported standard.",
			},
		},
	}
}
