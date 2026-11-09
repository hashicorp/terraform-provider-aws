// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	canonicalBucketNamePatternNoAnchors = `[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]`
	canonicalBucketNamePattern          = `^` + canonicalBucketNamePatternNoAnchors + `$`
)

// S3BucketName returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid S3 bucket name.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
var S3BucketName validator.String = stringvalidator.RegexMatches(
	regexache.MustCompile(canonicalBucketNamePattern),
	`Bucket names must be 3 to 63 characters and begin and end with a letter or number. Valid characters are a-z, 0-9, periods (.), and hyphens.`,
)
