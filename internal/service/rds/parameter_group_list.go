// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_db_parameter_group")
func newParameterGroupResourceAsListResource() inttypes.ListResourceForSDK {
	l := parameterGroupListResource{}
	l.SetResourceSchema(resourceParameterGroup())
	return &l
}

var _ list.ListResource = &parameterGroupListResource{}

type parameterGroupListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type parameterGroupListResourceModel struct {
	framework.WithRegionModel
}

func (l *parameterGroupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.RDSClient(ctx)

	var query parameterGroupListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing RDS DB Parameter Groups")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &rds.DescribeDBParameterGroupsInput{}
		for item, err := range listParameterGroups(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.DBParameterGroupName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				if err := resourceParameterGroupFlatten(ctx, conn, &item, rd); err != nil {
					if retry.NotFound(err) {
						tflog.Info(ctx, "Skipping RDS DB Parameter Group, not found when fetching parameters", map[string]any{
							"error": err.Error(),
						})
						continue
					}

					result := fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}
			}

			result.DisplayName = name

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

func listParameterGroups(ctx context.Context, conn *rds.Client, input *rds.DescribeDBParameterGroupsInput) iter.Seq2[awstypes.DBParameterGroup, error] {
	return func(yield func(awstypes.DBParameterGroup, error) bool) {
		pages := rds.NewDescribeDBParameterGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.DBParameterGroup{}, fmt.Errorf("listing RDS DB Parameter Group resources: %w", err))
				return
			}

			for _, item := range page.DBParameterGroups {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
