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
// @SDKListResource("aws_db_instance")
func newInstanceResourceAsListResource() inttypes.ListResourceForSDK {
	l := instanceListResource{}
	l.SetResourceSchema(resourceInstance())
	return &l
}

var _ list.ListResource = &instanceListResource{}

type instanceListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *instanceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.RDSClient(ctx)

	tflog.Info(ctx, "Listing RDS DB Instances")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &rds.DescribeDBInstancesInput{}
		for item, err := range listDBInstances(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			identifier := aws.ToString(item.DBInstanceIdentifier)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrIdentifier), identifier)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(aws.ToString(item.DbiResourceId))

			if request.IncludeResource {
				if diags := resourceInstanceFlatten(ctx, awsClient, &item, rd); diags.HasError() {
					tflog.Error(ctx, "Flattening RDS DB Instance", map[string]any{
						"error": diags[0].Summary,
					})
					continue
				}
			}

			result.DisplayName = identifier

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

func listDBInstances(ctx context.Context, conn *rds.Client, input *rds.DescribeDBInstancesInput) iter.Seq2[awstypes.DBInstance, error] {
	return func(yield func(awstypes.DBInstance, error) bool) {
		pages := rds.NewDescribeDBInstancesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.DBInstance{}, fmt.Errorf("listing RDS DB Instance resources: %w", err))
				return
			}

			for _, item := range page.DBInstances {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
