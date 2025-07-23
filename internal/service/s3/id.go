// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"fmt"
	"strings"
)

const resourceIDSeparator = ","

// createResourceID is a generic method for creating an ID string for a bucket-related resource e.g. aws_s3_bucket_versioning.
// The method expects a bucket name and an optional accountID.
func createResourceID(bucket, expectedBucketOwner string) string {
	if expectedBucketOwner == "" {
		return bucket
	}

	parts := []string{bucket, expectedBucketOwner}
	id := strings.Join(parts, resourceIDSeparator)

	return id
}

// parseResourceID is a generic method for parsing an ID string
// for a bucket name and accountID if provided.
func parseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, resourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", nil
	}

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected BUCKET or BUCKET%[2]sEXPECTED_BUCKET_OWNER", id, resourceIDSeparator)
}
