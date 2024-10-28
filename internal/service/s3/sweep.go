// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_s3_object", &resource.Sweeper{
		Name: "aws_s3_object",
		F:    sweepObjects,
		Dependencies: []string{
			"aws_m2_application",
		},
	})

	awsv2.Register("aws_s3_bucket", sweepBuckets,
		"aws_s3_access_point",
		"aws_s3_object",
		"aws_s3control_access_grants_instance",
		"aws_s3control_multi_region_access_point",
	)

	resource.AddTestSweepers("aws_s3_directory_bucket", &resource.Sweeper{
		Name: "aws_s3_directory_bucket",
		F:    sweepDirectoryBuckets,
		Dependencies: []string{
			"aws_s3_object",
		},
	})
}

func sweepObjects(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	sweepables, err := sweepGeneralPurposeBucketObjects(ctx, client)
	if awsv2.SkipSweepError(err) {
		tflog.Warn(ctx, "Skipping sweeper", map[string]any{
			"error": err.Error(),
		})
		return nil
	}
	if err != nil {
		return fmt.Errorf("error listing General Purpose S3 Buckets (%s): %w", region, err)
	}

	// Directory buckets.
	dbSweepables, err := sweepDirectoryBucketObjects(ctx, client)
	if awsv2.SkipSweepError(err) {
		tflog.Warn(ctx, "Skipping sweeper", map[string]any{
			"error": err.Error(),
		})
		// Allow objects in general purpose buckets to be deleted.
	}
	if err != nil {
		return fmt.Errorf("error listing S3 Directory Buckets (%s): %w", region, err)
	}
	sweepables = append(sweepables, dbSweepables...)

	err = sweep.SweepOrchestrator(ctx, sweepables)

	if err != nil {
		return fmt.Errorf("error sweeping S3 Objects (%s): %w", region, err)
	}

	return nil
}

func sweepGeneralPurposeBucketObjects(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3Client(ctx)

	var sweepables []sweep.Sweepable

	input := s3.ListBucketsInput{
		BucketRegion: aws.String(client.Region),
	}
	pages := s3.NewListBucketsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, bucket := range page.Buckets {
			bucketName := aws.ToString(bucket.Name)
			tflog.SetField(ctx, "bucket_name", bucketName)
			if bucketNameFilter(ctx, bucket) {
				var objectLockEnabled bool
				objLockConfig, err := findObjectLockConfiguration(ctx, conn, bucketName, "")
				if !tfresource.NotFound(err) {
					if err != nil {
						tflog.Warn(ctx, "Reading S3 Bucket Object Lock Configuration", map[string]any{
							"error": err.Error(),
						})
						continue
					}
					objectLockEnabled = objLockConfig.ObjectLockEnabled == types.ObjectLockEnabledEnabled
				}

				sweepables = append(sweepables, objectSweeper{
					conn:   conn,
					bucket: bucketName,
					locked: objectLockEnabled,
				})
			}
		}
	}

	return sweepables, nil
}

func sweepDirectoryBucketObjects(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3ExpressClient(ctx)

	var sweepables []sweep.Sweepable

	pages := s3.NewListDirectoryBucketsPaginator(conn, &s3.ListDirectoryBucketsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, bucket := range page.Buckets {
			bucketName := aws.ToString(bucket.Name)
			tflog.SetField(ctx, "bucket_name", bucketName)
			if bucketNameFilter(ctx, bucket) {
				sweepables = append(sweepables, directoryBucketObjectSweeper{
					conn:   conn,
					bucket: bucketName,
				})
			}
		}
	}

	return sweepables, nil
}

type objectSweeper struct {
	conn   *s3.Client
	bucket string
	locked bool
}

func (os objectSweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	// Delete everything including locked objects.
	log.Printf("[INFO] Emptying S3 Bucket (%s)", os.bucket)
	n, err := emptyBucket(ctx, os.conn, os.bucket, os.locked)
	if err != nil {
		return fmt.Errorf("deleting S3 Bucket (%s) objects: %w", os.bucket, err)
	}
	log.Printf("[INFO] Deleted %d S3 Objects from S3 Bucket (%s)", n, os.bucket)
	return nil
}

type directoryBucketObjectSweeper struct {
	conn   *s3.Client
	bucket string
}

func (os directoryBucketObjectSweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	log.Printf("[INFO] Emptying S3 Directory Bucket (%s)", os.bucket)
	n, err := emptyDirectoryBucket(ctx, os.conn, os.bucket)
	if err != nil {
		return fmt.Errorf("deleting S3 Directory Bucket (%s) objects: %w", os.bucket, err)
	}
	log.Printf("[INFO] Deleted %d S3 Objects from S3 Directory Bucket (%s)", n, os.bucket)
	return nil
}

func sweepBuckets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3Client(ctx)

	var sweepResources []sweep.Sweepable
	r := resourceBucket()

	input := s3.ListBucketsInput{
		BucketRegion: aws.String(client.Region),
	}
	pages := s3.NewListBucketsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, bucket := range page.Buckets {
			tflog.SetField(ctx, "bucket_name", aws.ToString(bucket.Name))
			if bucketNameFilter(ctx, bucket) {
				d := r.Data(nil)
				d.SetId(aws.ToString(bucket.Name))

				sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
			}
		}
	}

	return sweepResources, nil
}

func bucketNameFilter(ctx context.Context, bucket types.Bucket) bool {
	name := aws.ToString(bucket.Name)

	prefixes := []string{
		"tf-acc",
		"tf-object-test",
		"tf-test",
		"tftest.applicationversion",
		"terraform-remote-s3-test",
		"aws-security-data-lake-", // Orphaned by aws_securitylake_data_lake.
		"resource-test-terraform",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	defaultNameRegexp := regexache.MustCompile(fmt.Sprintf(`^%s\d+$`, id.UniqueIdPrefix))
	if defaultNameRegexp.MatchString(name) {
		return true
	}

	tflog.Info(ctx, "Skipping resource", map[string]any{
		"skip_reason": "no match on prefix list",
	})
	return false
}

func sweepDirectoryBuckets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.S3ExpressClient(ctx)
	input := &s3.ListDirectoryBucketsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := s3.NewListDirectoryBucketsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping S3 Directory Bucket sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing S3 Directory Buckets (%s): %w", region, err)
		}

		for _, v := range page.Buckets {
			if !bucketNameFilter(ctx, v) {
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newDirectoryBucketResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Name)),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping S3 Directory Buckets (%s): %w", region, err)
	}

	return nil
}
