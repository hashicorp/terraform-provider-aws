// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const resourceIDSeparator = ","

// createResourceID is a generic method for creating an ID string for a bucket-related resource e.g. aws_s3_bucket_versioning.
// The method expects a bucket name and an optional accountID.
func createResourceID(bucket, expectedBucketOwner string) string {
	if expectedBucketOwner == "" {
		return bucket
	}

	parts := []string{bucket, expectedBucketOwner}
	return strings.Join(parts, resourceIDSeparator)
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

var _ inttypes.SDKv2ImportID = resourceImportID{}

// resourceImportID is a generic custom import type supporting bucket-related resources.
//
// Resources which expect a bucket name and an optional accountID for the identifier
// can use this custom importer when adding resource identity support.
type resourceImportID struct{}

func (resourceImportID) Create(d *schema.ResourceData) string {
	bucket := d.Get(names.AttrBucket).(string)
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)
	return createResourceID(bucket, expectedBucketOwner)
}

func (resourceImportID) Parse(id string) (string, map[string]any, error) {
	bucket, expectedBucketOwner, err := parseResourceID(id)
	if err != nil {
		return id, nil, err
	}

	results := map[string]any{
		names.AttrBucket: bucket,
	}
	if expectedBucketOwner != "" {
		results[names.AttrExpectedBucketOwner] = expectedBucketOwner
	}

	return id, results, nil
}
