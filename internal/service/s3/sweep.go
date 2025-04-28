// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_s3_bucket", sweepBuckets,
		"aws_s3_access_point",
		"aws_s3_object_gp_bucket",
		"aws_s3control_access_grants_instance",
		"aws_s3control_multi_region_access_point",
	)

	awsv2.Register("aws_s3_directory_bucket", sweepDirectoryBuckets)

	awsv2.Register("aws_s3_object", sweepObjects)

	awsv2.Register("aws_s3_object_directory_bucket", sweepDirectoryBucketObjects)

	awsv2.Register("aws_s3_object_gp_bucket", sweepGeneralPurposeBucketObjects,
		"aws_m2_application",
	)
}

const logKeyBucketName = "bucket_name"

func sweepObjects(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	tflog.Info(ctx, "Noop sweeper")
	return nil, nil
}

func sweepGeneralPurposeBucketObjects(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3Client(ctx)

	var sweepables []sweep.Sweepable

	input := s3.ListBucketsInput{
		BucketRegion: aws.String(client.Region(ctx)),
	}
	pages := s3.NewListBucketsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, bucket := range page.Buckets {
			bucketName := aws.ToString(bucket.Name)
			ctx = tflog.SetField(ctx, logKeyBucketName, bucketName)
			if !bucketNameFilter(ctx, bucket) {
				continue
			}

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
			ctx = tflog.SetField(ctx, logKeyBucketName, bucketName)
			if !bucketNameFilter(ctx, bucket) {
				continue
			}

			sweepables = append(sweepables, directoryBucketObjectSweeper{
				conn:   conn,
				bucket: bucketName,
			})
		}
	}

	return sweepables, nil
}

type objectSweeper struct {
	conn   *s3.Client
	bucket string
	locked bool
}

func (os objectSweeper) Delete(ctx context.Context, optFns ...tfresource.OptionsFunc) error {
	// Delete everything including locked objects.
	tflog.Info(ctx, "Emptying S3 General Purpose Bucket")
	n, err := emptyBucket(ctx, os.conn, os.bucket, os.locked)
	if err != nil {
		return fmt.Errorf("deleting S3 Bucket (%s) objects: %w", os.bucket, err)
	}
	tflog.Info(ctx, "Deleted Objects from S3 General Purpose Bucket", map[string]any{
		"object_count": n,
	})
	return nil
}

type directoryBucketObjectSweeper struct {
	conn   *s3.Client
	bucket string
}

func (os directoryBucketObjectSweeper) Delete(ctx context.Context, optFns ...tfresource.OptionsFunc) error {
	tflog.Info(ctx, "Emptying S3 Directory Bucket")
	n, err := emptyDirectoryBucket(ctx, os.conn, os.bucket)
	if err != nil {
		return fmt.Errorf("deleting S3 Directory Bucket (%s) objects: %w", os.bucket, err)
	}
	tflog.Info(ctx, "Deleted Objects from S3 Directory Bucket", map[string]any{
		"object_count": n,
	})
	return nil
}

func sweepBuckets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3Client(ctx)

	var sweepResources []sweep.Sweepable
	r := resourceBucket()

	input := s3.ListBucketsInput{
		BucketRegion: aws.String(client.Region(ctx)),
	}
	pages := s3.NewListBucketsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, bucket := range page.Buckets {
			ctx = tflog.SetField(ctx, logKeyBucketName, aws.ToString(bucket.Name))
			if !bucketNameFilter(ctx, bucket) {
				continue
			}

			d := r.Data(nil)
			d.SetId(aws.ToString(bucket.Name))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
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

func sweepDirectoryBuckets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3ExpressClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := s3.NewListDirectoryBucketsPaginator(conn, &s3.ListDirectoryBucketsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, bucket := range page.Buckets {
			ctx = tflog.SetField(ctx, logKeyBucketName, aws.ToString(bucket.Name))
			if !bucketNameFilter(ctx, bucket) {
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newDirectoryBucketResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(bucket.Name)),
			))
		}
	}

	return sweepResources, nil
}
