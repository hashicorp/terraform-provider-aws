package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
)

// EmptyBucket empties the specified S3 bucket by deleting all object versions and delete markers.
// If `force` is `true` then S3 Object Lock governance mode restrictions are bypassed and
// an attempt is made to remove any S3 Object Lock legal holds.
func EmptyBucket(ctx context.Context, conn *s3.S3, bucket string, force bool) error {
	err := forEachObjectVersionsPage(ctx, conn, bucket, func(ctx context.Context, conn *s3.S3, bucket string, page *s3.ListObjectVersionsOutput) error {
		return deletePageOfObjectVersions(ctx, conn, bucket, force, page)
	})

	if err != nil {
		return err
	}

	err = forEachObjectVersionsPage(ctx, conn, bucket, deletePageOfDeleteMarkers)

	if err != nil {
		return err
	}

	return nil
}

// forEachObjectVersionsPage calls the specified function for each page returned from the S3 ListObjectVersionsPages API.
func forEachObjectVersionsPage(ctx context.Context, conn *s3.S3, bucket string, fn func(ctx context.Context, conn *s3.S3, bucket string, page *s3.ListObjectVersionsOutput) error) error {
	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	}
	var lastErr error

	err := conn.ListObjectVersionsPagesWithContext(ctx, input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		if err := fn(ctx, conn, bucket, page); err != nil {
			lastErr = err

			return false
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("listing S3 Bucket (%s) object versions: %w", bucket, err)
	}

	if lastErr != nil {
		return lastErr
	}

	return nil
}

func deletePageOfObjectVersions(ctx context.Context, conn *s3.S3, bucket string, force bool, page *s3.ListObjectVersionsOutput) error {
	toDelete := make([]*s3.ObjectIdentifier, 0, len(page.Versions))
	for _, v := range page.Versions {
		toDelete = append(toDelete, &s3.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		})
	}

	if len(toDelete) == 0 {
		return nil
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3.Delete{
			Objects: toDelete,
			Quiet:   aws.Bool(true), // Only report errors.
		},
	}

	if force {
		input.BypassGovernanceRetention = aws.Bool(true)
	}

	output, err := conn.DeleteObjectsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting S3 Bucket (%s) objects: %w", bucket, err)
	}

	var deleteErrs *multierror.Error

	for _, v := range output.Errors {
		// Attempt to remove any legal hold on the object.
		if force && aws.StringValue(v.Code) == ErrCodeAccessDenied {
			_, err := conn.PutObjectLegalHold(&s3.PutObjectLegalHoldInput{
				Bucket:    aws.String(bucket),
				Key:       v.Key,
				VersionId: v.VersionId,
				LegalHold: &s3.ObjectLockLegalHold{
					Status: aws.String(s3.ObjectLockLegalHoldStatusOff),
				},
			})

			if err != nil {
				// Add the original error and the new error.
				deleteErrs = multierror.Append(deleteErrs, newObjectOpError(v, "deleting"))
				deleteErrs = multierror.Append(deleteErrs, newObjectOpError(v, "removing legal hold"))
			}
		} else {
			deleteErrs = multierror.Append(deleteErrs, newObjectOpError(v, "deleting"))
		}
	}

	return deleteErrs.ErrorOrNil()
}

func deletePageOfDeleteMarkers(ctx context.Context, conn *s3.S3, bucket string, page *s3.ListObjectVersionsOutput) error {
	toDelete := make([]*s3.ObjectIdentifier, 0, len(page.Versions))
	for _, v := range page.DeleteMarkers {
		toDelete = append(toDelete, &s3.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		})
	}

	if len(toDelete) == 0 {
		return nil
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3.Delete{
			Objects: toDelete,
			Quiet:   aws.Bool(true), // Only report errors.
		},
	}

	output, err := conn.DeleteObjectsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting S3 Bucket (%s) delete markers: %w", bucket, err)
	}

	var deleteErrs *multierror.Error

	for _, v := range output.Errors {
		deleteErrs = multierror.Append(deleteErrs, newObjectOpError(v, "deleting"))
	}

	return deleteErrs.ErrorOrNil()
}

func newObjectOpError(v *s3.Error, op string) error {
	if v == nil {
		return nil
	}

	key := aws.StringValue(v.Key)
	awsErr := awserr.New(aws.StringValue(v.Code), aws.StringValue(v.Message), nil)

	if v.VersionId == nil {
		return fmt.Errorf("%s S3 object (%s): %w", op, key, awsErr)
	}

	return fmt.Errorf("%s S3 object (%s) version (%s): %w", op, key, aws.StringValue(v.VersionId), awsErr)
}
