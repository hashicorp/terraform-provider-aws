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
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
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

	resource.AddTestSweepers("aws_s3_bucket", &resource.Sweeper{
		Name: "aws_s3_bucket",
		F:    sweepBuckets,
		Dependencies: []string{
			"aws_s3_access_point",
			"aws_s3_object",
			"aws_s3control_access_grants_instance",
			"aws_s3control_multi_region_access_point",
		},
	})

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
	conn := client.S3Client(ctx)

	// General purpose buckets.
	output, err := conn.ListBuckets(ctx, &s3.ListBucketsInput{})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Objects sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing S3 Buckets: %w", err)
	}

	buckets := tfslices.Filter(output.Buckets, bucketRegionFilter(ctx, conn, region, client.S3UsePathStyle(ctx)))
	buckets = tfslices.Filter(buckets, bucketNameFilter)
	sweepables := make([]sweep.Sweepable, 0)

	for _, bucket := range buckets {
		bucket := aws.ToString(bucket.Name)
		objLockConfig, err := findObjectLockConfiguration(ctx, conn, bucket, "")

		var objectLockEnabled bool

		if !tfresource.NotFound(err) {
			if err != nil {
				log.Printf("[WARN] Reading S3 Bucket Object Lock Configuration (%s): %s", bucket, err)
				continue
			}
			objectLockEnabled = objLockConfig.ObjectLockEnabled == types.ObjectLockEnabledEnabled
		}

		sweepables = append(sweepables, objectSweeper{
			conn:   conn,
			bucket: bucket,
			locked: objectLockEnabled,
		})
	}

	// Directory buckets.
	s3ExpressConn := client.S3ExpressClient(ctx)
	pages := s3.NewListDirectoryBucketsPaginator(s3ExpressConn, &s3.ListDirectoryBucketsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping S3 Objects sweep for %s: %s", region, err)
			break // Allow objects in general purpose buckets to be deleted.
		}

		if err != nil {
			return fmt.Errorf("error listing S3 Directory Buckets (%s): %w", region, err)
		}

		for _, v := range page.Buckets {
			if !bucketNameFilter(v) {
				continue
			}

			sweepables = append(sweepables, directoryBucketObjectSweeper{
				conn:   s3ExpressConn,
				bucket: aws.ToString(v.Name),
			})
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepables)

	if err != nil {
		return fmt.Errorf("error sweeping S3 Objects (%s): %w", region, err)
	}

	return nil
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

func sweepBuckets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.S3Client(ctx)
	input := &s3.ListBucketsInput{}

	output, err := conn.ListBuckets(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Buckets sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing S3 Buckets: %w", err)
	}

	if len(output.Buckets) == 0 {
		log.Print("[DEBUG] No S3 Buckets to sweep")
		return nil
	}

	buckets := tfslices.Filter(output.Buckets, bucketRegionFilter(ctx, conn, region, client.S3UsePathStyle(ctx)))
	buckets = tfslices.Filter(buckets, bucketNameFilter)
	sweepables := make([]sweep.Sweepable, 0)

	for _, bucket := range buckets {
		name := aws.ToString(bucket.Name)

		r := resourceBucket()
		d := r.Data(nil)
		d.SetId(name)

		sweepables = append(sweepables, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepables)

	if err != nil {
		return fmt.Errorf("error sweeping S3 Buckets (%s): %w", region, err)
	}

	return nil
}

func bucketRegion(ctx context.Context, conn *s3.Client, bucket string, s3UsePathStyle bool) (string, error) {
	region, err := manager.GetBucketRegion(ctx, conn, bucket, func(o *s3.Options) {
		// By default, GetBucketRegion forces virtual host addressing, which
		// is not compatible with many non-AWS implementations. Instead, pass
		// the provider s3_force_path_style configuration, which defaults to
		// false, but allows override.
		o.UsePathStyle = s3UsePathStyle
	})

	if err != nil {
		return "", err
	}

	return region, nil
}

func bucketNameFilter(bucket types.Bucket) bool {
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

	if defaultNameRegexp.MatchString(name) {
		return true
	}

	log.Printf("[INFO] Skipping S3 Bucket (%s): not in prefix list", name)
	return false
}

var (
	defaultNameRegexp = regexache.MustCompile(fmt.Sprintf(`^%s\d+$`, id.UniqueIdPrefix))
)

func bucketRegionFilter(ctx context.Context, conn *s3.Client, region string, s3UsePathStyle bool) tfslices.Predicate[types.Bucket] {
	return func(bucket types.Bucket) bool {
		name := aws.ToString(bucket.Name)

		bucketRegion, err := bucketRegion(ctx, conn, name, s3UsePathStyle)

		if err != nil {
			log.Printf("[WARN] Getting S3 Bucket (%s) region: %s", name, err)
			return false
		}

		if bucketRegion != region {
			log.Printf("[INFO] Skipping S3 Bucket (%s): not in %s", name, region)
			return false
		}

		return true
	}
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
			if !bucketNameFilter(v) {
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
