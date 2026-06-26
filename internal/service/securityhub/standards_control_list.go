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
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_securityhub_standards_control")
func newStandardsControlResourceAsListResource() inttypes.ListResourceForSDK {
	l := standardsControlListResource{}
	l.SetResourceSchema(resourceStandardsControl())
	return &l
}

var _ list.ListResource = &standardsControlListResource{}

type standardsControlListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *standardsControlListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"standards_subscription_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN that represents your subscription to a supported standard.",
			},
		},
	}
}

func (l *standardsControlListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SecurityHubClient(ctx)

	var query listStandardsControlModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	standardsSubscriptionARN := fwflex.StringValueFromFramework(ctx, query.StandardsSubscriptionARN)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("standards_subscription_arn"): standardsSubscriptionARN,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := securityhub.DescribeStandardsControlsInput{
			StandardsSubscriptionArn: aws.String(standardsSubscriptionARN),
		}
		for item, err := range listStandardsControls(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.StandardsControlArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(arn)
			rd.Set("standards_control_arn", arn)

			if request.IncludeResource {
				if err := resourceStandardsControlFlatten(ctx, &item, rd); err != nil {
					tflog.Error(ctx, "Reading Security Hub Standards Control", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.Title)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
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
