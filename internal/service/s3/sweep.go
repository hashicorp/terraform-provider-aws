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
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_s3_bucket_object", &resource.Sweeper{
		Name: "aws_s3_bucket_object",
		F:    sweepBucketObjects,
	})

	resource.AddTestSweepers("aws_s3_bucket", &resource.Sweeper{
		Name: "aws_s3_bucket",
		F:    sweepBuckets,
		Dependencies: []string{
			"aws_s3_access_point",
			"aws_s3_bucket_object",
			"aws_s3control_multi_region_access_point",
		},
	})
}

func sweepBucketObjects(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).S3ConnURICleaningDisabled
	input := &s3.ListBucketsInput{}

	output, err := conn.ListBuckets(input)

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Bucket Objects sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing S3 Bucket Objects: %s", err)
	}

	if len(output.Buckets) == 0 {
		log.Print("[DEBUG] No S3 Bucket Objects to sweep")
		return nil
	}

	for _, bucket := range output.Buckets {
		bucketName := aws.StringValue(bucket.Name)

		hasPrefix := false
		prefixes := []string{"tf-acc", "tf-object-test", "tf-test", "tf-emr-bootstrap"}

		for _, prefix := range prefixes {
			if strings.HasPrefix(bucketName, prefix) {
				hasPrefix = true
				break
			}
		}

		if !hasPrefix {
			log.Printf("[INFO] Skipping S3 Bucket: %s", bucketName)
			continue
		}

		bucketRegion, err := bucketRegion(conn, bucketName)

		if err != nil {
			log.Printf("[ERROR] Error getting S3 Bucket (%s) Location: %s", bucketName, err)
			continue
		}

		if bucketRegion != region {
			log.Printf("[INFO] Skipping S3 Bucket (%s) in different region: %s", bucketName, bucketRegion)
			continue
		}

		objectLockEnabled, err := bucketObjectLockEnabled(conn, bucketName)

		if err != nil {
			log.Printf("[ERROR] Error getting S3 Bucket (%s) Object Lock: %s", bucketName, err)
			continue
		}

		// Delete everything including locked objects. Ignore any object errors.
		err = DeleteAllObjectVersions(conn, bucketName, "", objectLockEnabled, true)

		if err != nil {
			return fmt.Errorf("error listing S3 Bucket (%s) Objects: %s", bucketName, err)
		}
	}

	return nil
}

func sweepBuckets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).S3Conn
	input := &s3.ListBucketsInput{}

	output, err := conn.ListBuckets(input)

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Buckets sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing S3 Buckets: %s", err)
	}

	if len(output.Buckets) == 0 {
		log.Print("[DEBUG] No S3 Buckets to sweep")
		return nil
	}

	defaultNameRegexp := regexp.MustCompile(`^terraform-\d+$`)
	for _, bucket := range output.Buckets {
		name := aws.StringValue(bucket.Name)

		sweepable := false
		prefixes := []string{"tf-acc", "tf-object-test", "tf-test", "tf-emr-bootstrap", "terraform-remote-s3-test"}

		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				sweepable = true
				break
			}
		}

		if defaultNameRegexp.MatchString(name) {
			sweepable = true
		}

		if !sweepable {
			log.Printf("[INFO] Skipping S3 Bucket: %s", name)
			continue
		}

		bucketRegion, err := bucketRegion(conn, name)

		if err != nil {
			log.Printf("[ERROR] Error getting S3 Bucket (%s) Location: %s", name, err)
			continue
		}

		if bucketRegion != region {
			log.Printf("[INFO] Skipping S3 Bucket (%s) in different Region: %s", name, bucketRegion)
			continue
		}

		input := &s3.DeleteBucketInput{
			Bucket: bucket.Name,
		}

		log.Printf("[INFO] Deleting S3 Bucket: %s", name)
		err = resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.DeleteBucket(input)

			if tfawserr.ErrMessageContains(err, s3.ErrCodeNoSuchBucket, "") {
				return nil
			}

			if tfawserr.ErrMessageContains(err, "BucketNotEmpty", "") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("error deleting S3 Bucket (%s): %s", name, err)
		}
	}

	return nil
}

func bucketRegion(conn *s3.S3, bucket string) (string, error) {
	region, err := s3manager.GetBucketRegionWithClient(context.Background(), conn, bucket, func(r *request.Request) {
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

func bucketObjectLockEnabled(conn *s3.S3, bucket string) (bool, error) {
	input := &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.GetObjectLockConfiguration(input)

	if tfawserr.ErrMessageContains(err, "ObjectLockConfigurationNotFoundError", "") {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return aws.StringValue(output.ObjectLockConfiguration.ObjectLockEnabled) == s3.ObjectLockEnabledEnabled, nil
}
