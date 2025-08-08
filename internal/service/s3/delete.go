// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithy "github.com/aws/smithy-go"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// emptyBucket empties the specified S3 general purpose bucket by deleting all object versions and delete markers.
// If `force` is `true` then S3 Object Lock governance mode restrictions are bypassed and
// an attempt is made to remove any S3 Object Lock legal holds.
// Returns the number of object versions and delete markers deleted.
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

// emptyDirectoryBucket empties the specified S3 directory bucket by deleting all objects.
// Returns the number of objects deleted.
func emptyDirectoryBucket(ctx context.Context, conn *s3.Client, bucket string) (int64, error) {
	return forEachObjectsPage(ctx, conn, bucket, deletePageOfObjects)
}

// forEachObjectVersionsPage calls the specified function for each page returned from the S3 ListObjectVersionsPages API.
func forEachObjectVersionsPage(ctx context.Context, conn *s3.Client, bucket string, fn func(ctx context.Context, conn *s3.Client, bucket string, page *s3.ListObjectVersionsOutput) (int64, error)) (int64, error) {
	input := &s3.ListObjectVersionsInput{
		Bucket:       aws.String(bucket),
		EncodingType: types.EncodingTypeUrl,
	}
	var lastErr error
	var nObjects int64

	pages := s3.NewListObjectVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nObjects, fmt.Errorf("listing S3 bucket (%s) object versions: %w", bucket, err)
		}

		// Reverse URL-encoding from requested EncodingType: "url"
		for i, v := range page.Versions {
			key, err := url.QueryUnescape(aws.ToString(v.Key))
			if err != nil {
				return 0, fmt.Errorf("listing S3 bucket (%s) object versions: unescaping object version key: %w", bucket, err)
			}
			page.Versions[i].Key = aws.String(key)
		}
		for i, v := range page.DeleteMarkers {
			key, err := url.QueryUnescape(aws.ToString(v.Key))
			if err != nil {
				return 0, fmt.Errorf("listing S3 bucket (%s) object versions: unescaping delete marker key: %w", bucket, err)
			}
			page.DeleteMarkers[i].Key = aws.String(key)
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

// forEachObjectsPage calls the specified function for each page returned from the S3 ListObjectsV2 API.
func forEachObjectsPage(ctx context.Context, conn *s3.Client, bucket string, fn func(ctx context.Context, conn *s3.Client, bucket string, page *s3.ListObjectsV2Output) (int64, error)) (int64, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:       aws.String(bucket),
		EncodingType: types.EncodingTypeUrl,
	}
	var lastErr error
	var nObjects int64

	pages := s3.NewListObjectsV2Paginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nObjects, fmt.Errorf("listing S3 bucket (%s) objects: %w", bucket, err)
		}

		// Reverse URL-encoding from requested EncodingType: "url"
		for i, v := range page.Contents {
			key, err := url.QueryUnescape(aws.ToString(v.Key))
			if err != nil {
				return 0, fmt.Errorf("listing S3 bucket (%s) object: unescaping object key: %w", bucket, err)
			}
			page.Contents[i].Key = aws.String(key)
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
	toDelete := tfslices.ApplyToAll(page.Versions, func(v types.ObjectVersion) types.ObjectIdentifier {
		return types.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		}
	})

	return deletePage(ctx, conn, bucket, force, toDelete)
}

// deletePageOfDeleteMarkers deletes a page (<= 1000) of S3 object delete markers.
// Returns the number of delete markers deleted.
func deletePageOfDeleteMarkers(ctx context.Context, conn *s3.Client, bucket string, page *s3.ListObjectVersionsOutput) (int64, error) {
	toDelete := tfslices.ApplyToAll(page.DeleteMarkers, func(v types.DeleteMarkerEntry) types.ObjectIdentifier {
		return types.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		}
	})

	return deletePage(ctx, conn, bucket, false, toDelete)
}

// deletePageOfObjects deletes a page (<= 1000) of S3 objects.
// Returns the number of objects deleted.
func deletePageOfObjects(ctx context.Context, conn *s3.Client, bucket string, page *s3.ListObjectsV2Output) (int64, error) {
	toDelete := tfslices.ApplyToAll(page.Contents, func(v types.Object) types.ObjectIdentifier {
		return types.ObjectIdentifier{
			Key: v.Key,
		}
	})

	return deletePage(ctx, conn, bucket, false, toDelete)
}

func deletePage(ctx context.Context, conn *s3.Client, bucket string, force bool, toDelete []types.ObjectIdentifier) (int64, error) {
	if len(toDelete) == 0 {
		return 0, nil
	}

	// Typically every entry from toDelete ends up in toDeleteBulk, so pre-allocate
	toDeleteBulk := make([]types.ObjectIdentifier, 0, len(toDelete))
	var toDeleteSingly []types.ObjectIdentifier

	// Keys with characters outside the valid XML character set cannot be deleted by bulk DeleteObjects, and must be deleted individually
	// using DeleteObject.
	for _, v := range toDelete {
		if keyInXMLCharacterRange(aws.ToString(v.Key)) {
			toDeleteBulk = append(toDeleteBulk, v)
		} else {
			toDeleteSingly = append(toDeleteSingly, v)
		}
	}

	var nObjects int64
	var outputErrs []types.Error

	for _, v := range toDeleteSingly {
		key := aws.ToString(v.Key)
		versionID := aws.ToString(v.VersionId)

		err := deleteObjectVersion(ctx, conn, bucket, key, versionID, force)
		if err == nil {
			nObjects++
			continue
		}

		if err, ok := errs.As[smithy.APIError](err); ok {
			if errorCode := err.ErrorCode(); errorCode == errCodeNoSuchBucket {
				return int64(len(toDelete)), nil
			} else {
				outputErrs = append(outputErrs, types.Error{
					Code:      aws.String(errorCode),
					Key:       aws.String(key),
					VersionId: aws.String(versionID),
				})
			}
		} else {
			typeErr := fmt.Errorf("unexpected error type: %w", err)
			return nObjects, newObjectVersionError(key, versionID, typeErr)
		}
	}

	if len(toDeleteBulk) > 0 {
		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &types.Delete{
				Objects: toDeleteBulk,
				Quiet:   aws.Bool(true), // Only report errors.
			},
		}
		if force {
			input.BypassGovernanceRetention = aws.Bool(force)
		}

		output, err := conn.DeleteObjects(ctx, input)

		if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
			return int64(len(toDelete)), nil
		}

		if err != nil {
			return nObjects, fmt.Errorf("deleting S3 bucket (%s) object versions: %w", bucket, err)
		}

		nObjects += int64(len(toDeleteBulk)) - int64(len(output.Errors))
		outputErrs = append(outputErrs, output.Errors...)
	}

	var errs []error
	for _, v := range outputErrs {
		code := aws.ToString(v.Code)
		if code == errCodeNoSuchKey {
			continue
		}

		// Attempt to remove any legal hold on the object.
		if force && code == errCodeAccessDenied {
			key := aws.ToString(v.Key)
			versionID := aws.ToString(v.VersionId)
			input := &s3.PutObjectLegalHoldInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
				LegalHold: &types.ObjectLockLegalHold{
					Status: types.ObjectLockLegalHoldStatusOff,
				},
				VersionId: aws.String(versionID),
			}

			_, err := conn.PutObjectLegalHold(ctx, input)

			if err != nil {
				// Add the original error and the new error.
				errs = append(errs, newDeleteObjectVersionError(v))
				errs = append(errs, fmt.Errorf("removing legal hold: %w", newObjectVersionError(key, versionID, err)))
			} else {
				// Attempt to delete the object once the legal hold has been removed.
				input := &s3.DeleteObjectInput{
					Bucket:    aws.String(bucket),
					Key:       aws.String(key),
					VersionId: aws.String(versionID),
				}

				_, err := conn.DeleteObject(ctx, input)

				if err != nil {
					errs = append(errs, fmt.Errorf("deleting: %w", newObjectVersionError(key, versionID, err)))
				} else {
					nObjects++
				}
			}
		} else {
			errs = append(errs, newDeleteObjectVersionError(v))
		}
	}

	if err := errors.Join(errs...); err != nil {
		return nObjects, fmt.Errorf("deleting S3 bucket (%s) object versions: %w", bucket, err)
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

// deleteAllObjectVersions deletes all versions of a specified key from an S3 general purpose bucket.
// If key is empty then all versions of all objects are deleted.
// Set `force` to `true` to override any S3 object lock protections on object lock enabled buckets.
// Returns the number of objects deleted.
// Use `emptyBucket` to delete all versions of all objects in a bucket.
func deleteAllObjectVersions(ctx context.Context, conn *s3.Client, bucket, key string, force, ignoreObjectErrors bool, optFns ...func(*s3.Options)) (int64, error) {
	if key == "" {
		return 0, errors.New("use `emptyBucket` to delete all versions of all objects in an S3 general purpose bucket")
	}

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	}
	var lastErr error
	var nObjects int64

	pages := s3.NewListObjectVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx, optFns...)

		if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
			break
		}

		if err != nil {
			return nObjects, err
		}

		for _, objectVersion := range page.Versions {
			objectKey := aws.ToString(objectVersion.Key)
			objectVersionID := aws.ToString(objectVersion.VersionId)

			if key != objectKey {
				continue
			}

			err := deleteObjectVersion(ctx, conn, bucket, objectKey, objectVersionID, force, optFns...)

			if err == nil {
				nObjects++
			}

			if tfawserr.ErrCodeEquals(err, errCodeAccessDenied) && force {
				// Remove any legal hold.
				input := &s3.HeadObjectInput{
					Bucket:    aws.String(bucket),
					Key:       aws.String(objectKey),
					VersionId: aws.String(objectVersionID),
				}

				output, err := conn.HeadObject(ctx, input, optFns...)

				if err != nil {
					log.Printf("[ERROR] Getting S3 Bucket (%s) Object (%s) Version (%s) metadata: %s", bucket, objectKey, objectVersionID, err)
					lastErr = err
					continue
				}

				if output.ObjectLockLegalHoldStatus == types.ObjectLockLegalHoldStatusOn {
					input := &s3.PutObjectLegalHoldInput{
						Bucket: aws.String(bucket),
						Key:    aws.String(objectKey),
						LegalHold: &types.ObjectLockLegalHold{
							Status: types.ObjectLockLegalHoldStatusOff,
						},
						VersionId: aws.String(objectVersionID),
					}

					_, err := conn.PutObjectLegalHold(ctx, input, optFns...)

					if err != nil {
						log.Printf("[ERROR] Putting S3 Bucket (%s) Object (%s) Version(%s) legal hold: %s", bucket, objectKey, objectVersionID, err)
						lastErr = err
						continue
					}

					// Attempt to delete again.
					err = deleteObjectVersion(ctx, conn, bucket, objectKey, objectVersionID, force, optFns...)

					if err != nil {
						lastErr = err
					} else {
						nObjects++
					}

					continue
				}

				// AccessDenied for another reason.
				lastErr = fmt.Errorf("deleting S3 Bucket (%s) Object (%s) Version (%s): %w", bucket, objectKey, objectVersionID, err)
				continue
			}

			if err != nil {
				lastErr = err
			}
		}
	}

	if lastErr != nil {
		if !ignoreObjectErrors {
			return nObjects, fmt.Errorf("deleting at least one S3 Object version, last error: %w", lastErr)
		}

		lastErr = nil
	}

	pages = s3.NewListObjectVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx, optFns...)

		if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
			break
		}

		if err != nil {
			return nObjects, err
		}

		for _, deleteMarker := range page.DeleteMarkers {
			deleteMarkerKey := aws.ToString(deleteMarker.Key)
			deleteMarkerVersionID := aws.ToString(deleteMarker.VersionId)

			if key != deleteMarkerKey {
				continue
			}

			// Delete markers have no object lock protections.
			err := deleteObjectVersion(ctx, conn, bucket, deleteMarkerKey, deleteMarkerVersionID, false, optFns...)

			if err != nil {
				lastErr = err
			} else {
				nObjects++
			}
		}
	}

	if lastErr != nil {
		if !ignoreObjectErrors {
			return nObjects, fmt.Errorf("deleting at least one S3 Object delete marker, last error: %w", lastErr)
		}
	}

	return nObjects, nil
}

// deleteObjectVersion deletes a specific object version.
// Set `force` to `true` to override any S3 object lock protections.
func deleteObjectVersion(ctx context.Context, conn *s3.Client, b, k, v string, force bool, optFns ...func(*s3.Options)) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(k),
	}
	if v != "" {
		input.VersionId = aws.String(v)
	}
	if force {
		input.BypassGovernanceRetention = aws.Bool(force)
	}

	log.Printf("[INFO] Deleting S3 Bucket (%s) Object (%s) Version (%s)", b, k, v)
	_, err := conn.DeleteObject(ctx, input, optFns...)

	if err != nil {
		log.Printf("[WARN] Deleting S3 Bucket (%s) Object (%s) Version (%s): %s", b, k, v, err)
	}

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchKey) {
		return nil
	}

	return err
}

func keyInXMLCharacterRange(key string) bool {
	for _, r := range key {
		if !(r == 0x09 ||
			r == 0x0A ||
			r == 0x0D ||
			r >= 0x20 && r <= 0xD7FF ||
			r >= 0xE000 && r <= 0xFFFD ||
			r >= 0x10000 && r <= 0x10FFFF) {
			return false
		}
	}
	return true
}
