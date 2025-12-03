// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_s3tables_namespace", sweepNamespaces, "aws_s3tables_table")
	awsv2.Register("aws_s3tables_table", sweepTables)
	awsv2.Register("aws_s3tables_table_bucket", sweepTableBuckets, "aws_s3tables_namespace")
}

func sweepNamespaces(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3TablesClient(ctx)
	var input s3tables.ListTableBucketsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := s3tables.NewListTableBucketsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.TableBuckets {
			tableBucketARN := aws.ToString(v.Arn)

			if typ := v.Type; typ != awstypes.TableBucketTypeCustomer {
				log.Printf("[INFO] Skipping S3 Tables Table Bucket %s: Type=%s", tableBucketARN, typ)
				continue
			}

			input := s3tables.ListNamespacesInput{
				TableBucketARN: aws.String(tableBucketARN),
			}
			pages := s3tables.NewListNamespacesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					return nil, err
				}

				for _, v := range page.Namespaces {
					sweepResources = append(sweepResources, framework.NewSweepResource(newNamespaceResource, client,
						framework.NewAttribute("table_bucket_arn", tableBucketARN),
						framework.NewAttribute(names.AttrNamespace, v.Namespace[0]),
					))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepTables(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3TablesClient(ctx)
	var input s3tables.ListTableBucketsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := s3tables.NewListTableBucketsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.TableBuckets {
			tableBucketARN := aws.ToString(v.Arn)

			if typ := v.Type; typ != awstypes.TableBucketTypeCustomer {
				log.Printf("[INFO] Skipping S3 Tables Table Bucket %s: Type=%s", tableBucketARN, typ)
				continue
			}

			input := s3tables.ListNamespacesInput{
				TableBucketARN: aws.String(tableBucketARN),
			}
			pages := s3tables.NewListNamespacesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					return nil, err
				}

				for _, v := range page.Namespaces {
					namespace := v.Namespace[0]
					input := s3tables.ListTablesInput{
						Namespace:      aws.String(namespace),
						TableBucketARN: aws.String(tableBucketARN),
					}
					pages := s3tables.NewListTablesPaginator(conn, &input)
					for pages.HasMorePages() {
						page, err := pages.NextPage(ctx)

						if err != nil {
							return nil, err
						}

						for _, v := range page.Tables {
							sweepResources = append(sweepResources, framework.NewSweepResource(newTableResource, client,
								framework.NewAttribute("table_bucket_arn", tableBucketARN),
								framework.NewAttribute(names.AttrNamespace, namespace),
								framework.NewAttribute(names.AttrName, aws.ToString(v.Name)),
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
	var input s3tables.ListTableBucketsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := s3tables.NewListTableBucketsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.TableBuckets {
			tableBucketARN := aws.ToString(v.Arn)

			if typ := v.Type; typ != awstypes.TableBucketTypeCustomer {
				log.Printf("[INFO] Skipping S3 Tables Table Bucket %s: Type=%s", tableBucketARN, typ)
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newTableBucketResource, client,
				framework.NewAttribute(names.AttrARN, tableBucketARN),
			))
		}
	}

	return sweepResources, nil
}
