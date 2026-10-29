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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_db_instance")
func newInstanceResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceInstance(), framework.SDKListSpec[awstypes.DBInstance, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.DBInstance, error], diag.Diagnostics) {
			var input rds.DescribeDBInstancesInput

			return listDBInstances(ctx, c.RDSClient(ctx), &input), nil
		},
		// The resource ID is immutable, but identity and display use the
		// user-facing identifier.
		ID: func(v awstypes.DBInstance, _ framework.RegionQuery) string {
			return aws.ToString(v.DbiResourceId)
		},
		Flatten:     resourceInstanceFlatten,
		DisplayName: func(v awstypes.DBInstance) string { return aws.ToString(v.DBInstanceIdentifier) },
	})
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
