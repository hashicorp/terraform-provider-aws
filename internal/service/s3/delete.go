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
// Returns the number of objects deleted.
func EmptyBucket(ctx context.Context, conn *s3.S3, bucket string, force bool) (int64, error) {
	nObjects, err := forEachObjectVersionsPage(ctx, conn, bucket, func(ctx context.Context, conn *s3.S3, bucket string, page *s3.ListObjectVersionsOutput) (int64, error) {
		return deletePageOfObjectVersions(ctx, conn, bucket, force, page)
	})

	if err != nil {
		return nObjects, err
	}

	n, err := forEachObjectVersionsPage(ctx, conn, bucket, deletePageOfDeleteMarkers)
	nObjects += n

	if err != nil {
		return nObjects, err
	}

	return nObjects, nil
}

// forEachObjectVersionsPage calls the specified function for each page returned from the S3 ListObjectVersionsPages API.
func forEachObjectVersionsPage(ctx context.Context, conn *s3.S3, bucket string, fn func(ctx context.Context, conn *s3.S3, bucket string, page *s3.ListObjectVersionsOutput) (int64, error)) (int64, error) {
	var nObjects int64

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	}
	var lastErr error

	err := conn.ListObjectVersionsPagesWithContext(ctx, input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		n, err := fn(ctx, conn, bucket, page)
		nObjects += n

		if err != nil {
			lastErr = err

			return false
		}

		return !lastPage
	})

	if err != nil {
		return nObjects, fmt.Errorf("listing S3 Bucket (%s) object versions: %w", bucket, err)
	}

	if lastErr != nil {
		return nObjects, lastErr
	}

	return nObjects, nil
}

// deletePageOfObjectVersions deletes a page (<= 1000) of S3 object versions.
// If `force` is `true` then S3 Object Lock governance mode restrictions are bypassed and
// an attempt is made to remove any S3 Object Lock legal holds.
// Returns the number of objects deleted.
func deletePageOfObjectVersions(ctx context.Context, conn *s3.S3, bucket string, force bool, page *s3.ListObjectVersionsOutput) (int64, error) {
	var nObjects int64

	toDelete := make([]*s3.ObjectIdentifier, 0, len(page.Versions))
	for _, v := range page.Versions {
		toDelete = append(toDelete, &s3.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		})
	}

	if nObjects = int64(len(toDelete)); nObjects == 0 {
		return nObjects, nil
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
		return nObjects, nil
	}

	if err != nil {
		return nObjects, fmt.Errorf("deleting S3 Bucket (%s) objects: %w", bucket, err)
	}

	nObjects -= int64(len(output.Errors))

	var deleteErrs *multierror.Error

	for _, v := range output.Errors {
		code := aws.StringValue(v.Code)

		if code == s3.ErrCodeNoSuchKey {
			continue
		}

		// Attempt to remove any legal hold on the object.
		if force && code == ErrCodeAccessDenied {
			key := aws.StringValue(v.Key)
			versionID := aws.StringValue(v.VersionId)

			_, err := conn.PutObjectLegalHoldWithContext(ctx, &s3.PutObjectLegalHoldInput{
				Bucket:    aws.String(bucket),
				Key:       aws.String(key),
				VersionId: aws.String(versionID),
				LegalHold: &s3.ObjectLockLegalHold{
					Status: aws.String(s3.ObjectLockLegalHoldStatusOff),
				},
			})

			if err != nil {
				// Add the original error and the new error.
				deleteErrs = multierror.Append(deleteErrs, newDeleteObjectVersionError(v))
				deleteErrs = multierror.Append(deleteErrs, fmt.Errorf("removing legal hold: %w", newObjectVersionError(key, versionID, err)))
			} else {
				// Attempt to delete the object once the legal hold has been removed.
				_, err := conn.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
					Bucket:    aws.String(bucket),
					Key:       aws.String(key),
					VersionId: aws.String(versionID),
				})

				if err != nil {
					deleteErrs = multierror.Append(deleteErrs, fmt.Errorf("deleting: %w", newObjectVersionError(key, versionID, err)))
				} else {
					nObjects++
				}
			}
		} else {
			deleteErrs = multierror.Append(deleteErrs, newDeleteObjectVersionError(v))
		}
	}

	if err := deleteErrs.ErrorOrNil(); err != nil {
		return nObjects, fmt.Errorf("deleting S3 Bucket (%s) objects: %w", bucket, err)
	}

	return nObjects, nil
}

// deletePageOfDeleteMarkers deletes a page (<= 1000) of S3 object delete markers.
// Returns the number of delete markers deleted.
func deletePageOfDeleteMarkers(ctx context.Context, conn *s3.S3, bucket string, page *s3.ListObjectVersionsOutput) (int64, error) {
	var nObjects int64

	toDelete := make([]*s3.ObjectIdentifier, 0, len(page.Versions))
	for _, v := range page.DeleteMarkers {
		toDelete = append(toDelete, &s3.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		})
	}

	if nObjects = int64(len(toDelete)); nObjects == 0 {
		return nObjects, nil
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
		return nObjects, nil
	}

	if err != nil {
		return nObjects, fmt.Errorf("deleting S3 Bucket (%s) delete markers: %w", bucket, err)
	}

	nObjects -= int64(len(output.Errors))

	var deleteErrs *multierror.Error

	for _, v := range output.Errors {
		deleteErrs = multierror.Append(deleteErrs, newDeleteObjectVersionError(v))
	}

	if err := deleteErrs.ErrorOrNil(); err != nil {
		return nObjects, fmt.Errorf("deleting S3 Bucket (%s) delete markers: %w", bucket, err)
	}

	return nObjects, nil
}

func newObjectVersionError(key, versionID string, err error) error {
	if err == nil {
		return nil
	}

	if versionID == "" {
		return fmt.Errorf("S3 object (%s): %w", key, err)
	}

	return fmt.Errorf("S3 object (%s) version (%s): %w", key, versionID, err)
}

func newDeleteObjectVersionError(err *s3.Error) error {
	if err == nil {
		return nil
	}

	awsErr := awserr.New(aws.StringValue(err.Code), aws.StringValue(err.Message), nil)

	return fmt.Errorf("deleting: %w", newObjectVersionError(aws.StringValue(err.Key), aws.StringValue(err.VersionId), awsErr))
}
