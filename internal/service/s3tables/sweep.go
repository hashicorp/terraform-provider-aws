// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_s3tables_table_bucket", sweepTableBuckets)
}

func sweepTableBuckets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3TablesClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := s3tables.NewListTableBucketsPaginator(conn, &s3tables.ListTableBucketsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, bucket := range page.TableBuckets {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceTableBucket, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(bucket.Arn)),
			))
		}
	}

	return sweepResources, nil
}
