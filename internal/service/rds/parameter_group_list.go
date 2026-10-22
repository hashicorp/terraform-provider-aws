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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_db_parameter_group")
func newParameterGroupResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceParameterGroup(), framework.SDKListSpec[awstypes.DBParameterGroup, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.DBParameterGroup, error], diag.Diagnostics) {
			var input rds.DescribeDBParameterGroupsInput

			return listParameterGroups(ctx, c.RDSClient(ctx), &input), nil
		},
		ID: func(pg awstypes.DBParameterGroup, _ framework.RegionQuery) string {
			return aws.ToString(pg.DBParameterGroupName)
		},
		Prepare: func(_ context.Context, pg awstypes.DBParameterGroup, _ framework.RegionQuery, rd *schema.ResourceData) {
			// Identity is the name, and Flatten only runs when resource state
			// is asked for.
			rd.Set(names.AttrName, aws.ToString(pg.DBParameterGroupName))
		},
		Flatten: resourceParameterGroupFlatten,
		// Flatten fetches the group's parameters, so a listing that only wants
		// identity should not pay for that call.
		FlattenWhenRequested: true,
	})
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
