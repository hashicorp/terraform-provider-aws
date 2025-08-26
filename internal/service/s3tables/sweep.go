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
	awsv2.Register("aws_s3tables_namespace", sweepNamespaces,
		"aws_s3tables_table",
	)

	awsv2.Register("aws_s3tables_table", sweepTables)

	awsv2.Register("aws_s3tables_table_bucket", sweepTableBuckets,
		"aws_s3tables_namespace",
	)
}

func sweepNamespaces(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3TablesClient(ctx)

	var sweepResources []sweep.Sweepable

	tableBuckets := s3tables.NewListTableBucketsPaginator(conn, &s3tables.ListTableBucketsInput{})
	for tableBuckets.HasMorePages() {
		page, err := tableBuckets.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, bucket := range page.TableBuckets {
			namespaces := s3tables.NewListNamespacesPaginator(conn, &s3tables.ListNamespacesInput{
				TableBucketARN: bucket.Arn,
			})
			for namespaces.HasMorePages() {
				page, err := namespaces.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, namespace := range page.Namespaces {
					sweepResources = append(sweepResources, framework.NewSweepResource(newNamespaceResource, client,
						framework.NewAttribute("table_bucket_arn", aws.ToString(bucket.Arn)),
						framework.NewAttribute(names.AttrNamespace, namespace.Namespace[0]),
					))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepTables(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3TablesClient(ctx)

	var sweepResources []sweep.Sweepable

	tableBuckets := s3tables.NewListTableBucketsPaginator(conn, &s3tables.ListTableBucketsInput{})
	for tableBuckets.HasMorePages() {
		page, err := tableBuckets.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, bucket := range page.TableBuckets {
			namespaces := s3tables.NewListNamespacesPaginator(conn, &s3tables.ListNamespacesInput{
				TableBucketARN: bucket.Arn,
			})
			for namespaces.HasMorePages() {
				page, err := namespaces.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, namespace := range page.Namespaces {
					tables := s3tables.NewListTablesPaginator(conn, &s3tables.ListTablesInput{
						TableBucketARN: bucket.Arn,
						Namespace:      aws.String(namespace.Namespace[0]),
					})
					for tables.HasMorePages() {
						page, err := tables.NextPage(ctx)
						if err != nil {
							return nil, err
						}

						for _, table := range page.Tables {
							sweepResources = append(sweepResources, framework.NewSweepResource(newTableResource, client,
								framework.NewAttribute("table_bucket_arn", aws.ToString(bucket.Arn)),
								framework.NewAttribute(names.AttrNamespace, namespace.Namespace[0]),
								framework.NewAttribute(names.AttrName, aws.ToString(table.Name)),
							))
						}
					}
				}
			}
		}
	}

	return sweepResources, nil
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
			sweepResources = append(sweepResources, framework.NewSweepResource(newTableBucketResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(bucket.Arn)),
			))
		}
	}

	return sweepResources, nil
}
