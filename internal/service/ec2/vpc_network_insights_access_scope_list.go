// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_ec2_network_insights_access_scope")
func newNetworkInsightsAccessScopeResourceAsListResource() list.ListResourceWithConfigure {
	return &networkInsightsAccessScopeListResource{}
}

var _ list.ListResource = &networkInsightsAccessScopeListResource{}

type networkInsightsAccessScopeListResource struct {
	networkInsightsAccessScopeResource
	framework.WithList
}

func (l *networkInsightsAccessScopeListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := ec2.DescribeNetworkInsightsAccessScopesInput{}
		for item, err := range listNetworkInsightsAccessScopeItems(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(item.NetworkInsightsAccessScopeId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			tags := keyValueTags(ctx, item.Tags)
			if v, ok := tags["Name"]; ok {
				result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), id)
			} else {
				result.DisplayName = id
			}

			var data networkInsightsAccessScopeResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ID = fwflex.StringValueToFramework(ctx, id)

				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, &item, &data)...)
				}
			})

			if !yield(result) {
				return
			}
		}
	}
}

func (r *networkInsightsAccessScopeResource) flatten(ctx context.Context, scope *awstypes.NetworkInsightsAccessScope, data *networkInsightsAccessScopeResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, scope, data, fwflex.WithFieldNamePrefix("NetworkInsightsAccessScope"))...)
	if diags.HasError() {
		return diags
	}

	conn := r.Meta().EC2Client(ctx)

	id := aws.ToString(scope.NetworkInsightsAccessScopeId)
	contentInput := ec2.GetNetworkInsightsAccessScopeContentInput{
		NetworkInsightsAccessScopeId: aws.String(id),
	}
	contentOutput, err := conn.GetNetworkInsightsAccessScopeContent(ctx, &contentInput)
	if err != nil {
		diags.AddError(fmt.Sprintf("reading EC2 Network Insights Access Scope (%s) content", id), err.Error())
		return diags
	}

	if v := contentOutput.NetworkInsightsAccessScopeContent; v != nil {
		diags.Append(fwflex.Flatten(ctx, v, data)...)
	}

	setTagsOut(ctx, scope.Tags)

	return diags
}

func listNetworkInsightsAccessScopeItems(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInsightsAccessScopesInput) iter.Seq2[awstypes.NetworkInsightsAccessScope, error] {
	return func(yield func(awstypes.NetworkInsightsAccessScope, error) bool) {
		pages := ec2.NewDescribeNetworkInsightsAccessScopesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.NetworkInsightsAccessScope{}, fmt.Errorf("listing EC2 Network Insights Access Scopes: %w", err))
				return
			}

			for _, item := range page.NetworkInsightsAccessScopes {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
