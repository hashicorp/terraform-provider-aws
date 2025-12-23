// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package s3vectors

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_s3vectors_vector_bucket", sweepVectorBuckets, "aws_s3vectors_index")
	awsv2.Register("aws_s3vectors_index", sweepIndexes)
}

func sweepIndexes(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3VectorsClient(ctx)
	var input s3vectors.ListVectorBucketsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := s3vectors.NewListVectorBucketsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.VectorBuckets {
			input := s3vectors.ListIndexesInput{
				VectorBucketName: v.VectorBucketName,
			}

			pages := s3vectors.NewListIndexesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					return nil, err
				}

				for _, v := range page.Indexes {
					sweepResources = append(sweepResources, framework.NewSweepResource(newIndexResource, client,
						framework.NewAttribute("index_arn", aws.ToString(v.IndexArn))))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepVectorBuckets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3VectorsClient(ctx)
	var input s3vectors.ListVectorBucketsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := s3vectors.NewListVectorBucketsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.VectorBuckets {
			sweepResources = append(sweepResources, framework.NewSweepResource(newVectorBucketResource, client,
				framework.NewAttribute("vector_bucket_arn", aws.ToString(v.VectorBucketArn))))
		}
	}

	return sweepResources, nil
}
