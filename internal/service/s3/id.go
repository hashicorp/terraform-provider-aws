// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"fmt"
	"strings"
)

const resourceIDSeparator = ","

// CreateResourceID is a generic method for creating an ID string for a bucket-related resource e.g. aws_s3_bucket_versioning.
// The method expects a bucket name and an optional accountID.
func CreateResourceID(bucket, expectedBucketOwner string) string {
	if expectedBucketOwner == "" {
		return bucket
	}

	parts := []string{bucket, expectedBucketOwner}
	id := strings.Join(parts, resourceIDSeparator)

	return id
}

// ParseResourceID is a generic method for parsing an ID string
// for a bucket name and accountID if provided.
func ParseResourceID(id string) (bucket, expectedBucketOwner string, err error) {
	parts := strings.Split(id, resourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		bucket = parts[0]
		return
	}

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		bucket = parts[0]
		expectedBucketOwner = parts[1]
		return
	}

	err = fmt.Errorf("unexpected format for ID (%s), expected BUCKET or BUCKET%sEXPECTED_BUCKET_OWNER", id, resourceIDSeparator)
	return
}
