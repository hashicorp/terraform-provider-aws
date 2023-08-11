// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package s3

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_s3_object", &resource.Sweeper{
		Name: "aws_s3_object",
		F:    sweepObjects,
	})

	resource.AddTestSweepers("aws_s3_bucket", &resource.Sweeper{
		Name: "aws_s3_bucket",
		F:    sweepBuckets,
		Dependencies: []string{
			"aws_s3_access_point",
			"aws_s3_object",
			"aws_s3control_multi_region_access_point",
		},
	})
}

func sweepObjects(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.S3ConnURICleaningDisabled(ctx)
	input := &s3.ListBucketsInput{}

	output, err := conn.ListBucketsWithContext(ctx, input)
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Objects sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("listing S3 Buckets: %w", err)
	}

	if len(output.Buckets) == 0 {
		log.Print("[DEBUG] No S3 Objects to sweep")
		return nil
	}

	sweepables := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	buckets, err := filterBuckets(output.Buckets, bucketRegionFilter(ctx, conn, region))
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	buckets, err = filterBuckets(buckets, bucketNameFilter)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	for _, bucket := range buckets {
		bucketName := aws.StringValue(bucket.Name)

		objectLockEnabled, err := objectLockEnabled(ctx, conn, bucketName)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("reading S3 Bucket (%s) object lock: %w", bucketName, err))
			continue
		}

		sweepables = append(sweepables, objectSweeper{
			conn:   conn,
			name:   bucketName,
			locked: objectLockEnabled,
		})
	}

	if err := sweep.SweepOrchestrator(ctx, sweepables); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping DynamoDB Backups for %s: %w", region, err))
	}

	return errs.ErrorOrNil()
}

type objectSweeper struct {
	conn   *s3.S3
	name   string
	locked bool
}

func (os objectSweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	// Delete everything including locked objects
	_, err := DeleteAllObjectVersions(ctx, os.conn, os.name, "", os.locked, true)
	if err != nil {
		return fmt.Errorf("deleting S3 Bucket (%s) contents: %w", os.name, err)
	}
	return nil
}

func sweepBuckets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.S3Conn(ctx)
	input := &s3.ListBucketsInput{}

	output, err := conn.ListBucketsWithContext(ctx, input)

	if sweep.SkipSweepError(err) {
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

	var errs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	buckets, err := filterBuckets(output.Buckets, bucketRegionFilter(ctx, conn, region))
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	buckets, err = filterBuckets(buckets, bucketNameFilter)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	for _, bucket := range buckets {
		name := aws.StringValue(bucket.Name)

		r := ResourceBucket()
		d := r.Data(nil)
		d.SetId(name)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping S3 Buckets for %s: %w", region, err))
	}

	return errs.ErrorOrNil()
}

func bucketRegion(ctx context.Context, conn *s3.S3, bucket string) (string, error) {
	region, err := s3manager.GetBucketRegionWithClient(ctx, conn, bucket, func(r *request.Request) {
		// By default, GetBucketRegion forces virtual host addressing, which
		// is not compatible with many non-AWS implementations. Instead, pass
		// the provider s3_force_path_style configuration, which defaults to
		// false, but allows override.
		r.Config.S3ForcePathStyle = conn.Config.S3ForcePathStyle
	})
	if err != nil {
		return "", err
	}

	return region, nil
}

func objectLockEnabled(ctx context.Context, conn *s3.S3, bucket string) (bool, error) {
	output, err := FindObjectLockConfiguration(ctx, conn, bucket, "")

	if tfresource.NotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return aws.StringValue(output.ObjectLockEnabled) == s3.ObjectLockEnabledEnabled, nil
}

type bucketFilter func(*s3.Bucket) (bool, error)

func filterBuckets(in []*s3.Bucket, f bucketFilter) ([]*s3.Bucket, error) {
	var errs *multierror.Error
	var out []*s3.Bucket

	for _, b := range in {
		if ok, err := f(b); err != nil {
			errs = multierror.Append(errs, err)
		} else if ok {
			out = append(out, b)
		}
	}

	return out, errs.ErrorOrNil()
}

func bucketNameFilter(bucket *s3.Bucket) (bool, error) {
	name := aws.StringValue(bucket.Name)

	prefixes := []string{
		"tf-acc",
		"tf-object-test",
		"tf-test",
		"tftest.applicationversion",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true, nil
		}
	}

	if defaultNameRegexp.MatchString(name) {
		return true, nil
	}

	log.Printf("[INFO] Skipping S3 Bucket (%s): not in prefix list", name)
	return false, nil
}

var (
	defaultNameRegexp = regexp.MustCompile(fmt.Sprintf(`^%s\d+$`, id.UniqueIdPrefix))
)

func bucketRegionFilter(ctx context.Context, conn *s3.S3, region string) bucketFilter {
	return func(bucket *s3.Bucket) (bool, error) {
		name := aws.StringValue(bucket.Name)

		bucketRegion, err := bucketRegion(ctx, conn, name)
		if err != nil {
			return false, err
		}
		if bucketRegion != region {
			log.Printf("[INFO] Skipping S3 Bucket (%s): not in %s", name, region)
			return false, nil
		}

		return true, nil
	}
}
