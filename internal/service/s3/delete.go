// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
)

// emptyBucket empties the specified S3 bucket by deleting all object versions and delete markers.
// If `force` is `true` then S3 Object Lock governance mode restrictions are bypassed and
// an attempt is made to remove any S3 Object Lock legal holds.
// Returns the number of objects deleted.
func emptyBucket(ctx context.Context, conn *s3.Client, bucket string, force bool) (int64, error) {
	nObjects, err := forEachObjectVersionsPage(ctx, conn, bucket, func(ctx context.Context, conn *s3.Client, bucket string, page *s3.ListObjectVersionsOutput) (int64, error) {
		return deletePageOfObjectVersions(ctx, conn, bucket, force, page)
	})

	if err != nil {
		return nObjects, err
	}

	n, err := forEachObjectVersionsPage(ctx, conn, bucket, deletePageOfDeleteMarkers)
	nObjects += n

	return nObjects, err
}

// forEachObjectVersionsPage calls the specified function for each page returned from the S3 ListObjectVersionsPages API.
func forEachObjectVersionsPage(ctx context.Context, conn *s3.Client, bucket string, fn func(ctx context.Context, conn *s3.Client, bucket string, page *s3.ListObjectVersionsOutput) (int64, error)) (int64, error) {
	var nObjects int64

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	}
	var lastErr error

	pages := s3.NewListObjectVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nObjects, fmt.Errorf("listing S3 bucket (%s) object versions: %w", bucket, err)
		}

		n, err := fn(ctx, conn, bucket, page)
		nObjects += n

		if err != nil {
			lastErr = err
			break
		}
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
func deletePageOfObjectVersions(ctx context.Context, conn *s3.Client, bucket string, force bool, page *s3.ListObjectVersionsOutput) (int64, error) {
	var nObjects int64

	toDelete := make([]types.ObjectIdentifier, 0, len(page.Versions))
	for _, v := range page.Versions {
		toDelete = append(toDelete, types.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		})
	}

	if nObjects = int64(len(toDelete)); nObjects == 0 {
		return nObjects, nil
	}

	input := &s3.DeleteObjectsInput{
		Bucket:                    aws.String(bucket),
		BypassGovernanceRetention: force,
		Delete: &types.Delete{
			Objects: toDelete,
			Quiet:   true, // Only report errors.
		},
	}

	output, err := conn.DeleteObjects(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return nObjects, nil
	}

	if err != nil {
		return nObjects, fmt.Errorf("deleting S3 bucket (%s) objects: %w", bucket, err)
	}

	nObjects -= int64(len(output.Errors))

	var deleteErrs []error

	for _, v := range output.Errors {
		code := aws.ToString(v.Code)

		if code == errCodeNoSuchKey {
			continue
		}

		// Attempt to remove any legal hold on the object.
		if force && code == errCodeAccessDenied {
			key := aws.ToString(v.Key)
			versionID := aws.ToString(v.VersionId)

			_, err := conn.PutObjectLegalHold(ctx, &s3.PutObjectLegalHoldInput{
				Bucket:    aws.String(bucket),
				Key:       aws.String(key),
				VersionId: aws.String(versionID),
				LegalHold: &types.ObjectLockLegalHold{
					Status: types.ObjectLockLegalHoldStatusOff,
				},
			})

			if err != nil {
				// Add the original error and the new error.
				deleteErrs = append(deleteErrs, newDeleteObjectVersionError(v))
				deleteErrs = append(deleteErrs, fmt.Errorf("removing legal hold: %w", newObjectVersionError(key, versionID, err)))
			} else {
				// Attempt to delete the object once the legal hold has been removed.
				_, err := conn.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket:    aws.String(bucket),
					Key:       aws.String(key),
					VersionId: aws.String(versionID),
				})

				if err != nil {
					deleteErrs = append(deleteErrs, fmt.Errorf("deleting: %w", newObjectVersionError(key, versionID, err)))
				} else {
					nObjects++
				}
			}
		} else {
			deleteErrs = append(deleteErrs, newDeleteObjectVersionError(v))
		}
	}

	if err := errors.Join(deleteErrs...); err != nil {
		return nObjects, fmt.Errorf("deleting S3 bucket (%s) objects: %w", bucket, err)
	}

	return nObjects, nil
}

// deletePageOfDeleteMarkers deletes a page (<= 1000) of S3 object delete markers.
// Returns the number of delete markers deleted.
func deletePageOfDeleteMarkers(ctx context.Context, conn *s3.Client, bucket string, page *s3.ListObjectVersionsOutput) (int64, error) {
	var nObjects int64

	toDelete := make([]types.ObjectIdentifier, 0, len(page.Versions))
	for _, v := range page.DeleteMarkers {
		toDelete = append(toDelete, types.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		})
	}

	if nObjects = int64(len(toDelete)); nObjects == 0 {
		return nObjects, nil
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: toDelete,
			Quiet:   true, // Only report errors.
		},
	}

	output, err := conn.DeleteObjects(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return nObjects, nil
	}

	if err != nil {
		return nObjects, fmt.Errorf("deleting S3 bucket (%s) delete markers: %w", bucket, err)
	}

	nObjects -= int64(len(output.Errors))

	var deleteErrs []error

	for _, v := range output.Errors {
		deleteErrs = append(deleteErrs, newDeleteObjectVersionError(v))
	}

	if err := errors.Join(deleteErrs...); err != nil {
		return nObjects, fmt.Errorf("deleting S3 bucket (%s) delete markers: %w", bucket, err)
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

func newDeleteObjectVersionError(err types.Error) error {
	s3Err := fmt.Errorf("%s: %s", aws.ToString(err.Code), aws.ToString(err.Message))

	return fmt.Errorf("deleting: %w", newObjectVersionError(aws.ToString(err.Key), aws.ToString(err.VersionId), s3Err))
}
