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
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_db_subnet_group")
func newSubnetGroupResourceAsListResource() inttypes.ListResourceForSDK {
	l := subnetGroupListResource{}
	l.SetResourceSchema(resourceSubnetGroup())
	return &l
}

var _ list.ListResource = &subnetGroupListResource{}

type subnetGroupListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type subnetGroupListResourceModel struct {
	framework.WithRegionModel
}

func (l *subnetGroupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.RDSClient(ctx)

	var query subnetGroupListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing RDS DB Subnet Groups")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &rds.DescribeDBSubnetGroupsInput{}
		for item, err := range listSubnetGroups(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.DBSubnetGroupName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				if err := resourceSubnetGroupFlatten(&item, rd); err != nil {
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

func listSubnetGroups(ctx context.Context, conn *rds.Client, input *rds.DescribeDBSubnetGroupsInput) iter.Seq2[awstypes.DBSubnetGroup, error] {
	return func(yield func(awstypes.DBSubnetGroup, error) bool) {
		pages := rds.NewDescribeDBSubnetGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.DBSubnetGroup{}, fmt.Errorf("listing RDS DB Subnet Group resources: %w", err))
				return
			}

			for _, item := range page.DBSubnetGroups {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
