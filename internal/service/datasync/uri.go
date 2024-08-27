// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

var (
	locationURIPattern                      = regexache.MustCompile(`^(azure-blob|efs|fsx[0-9a-z-]+|hdfs|nfs|s3|smb)://(.+)$`)
	locationURIGlobalIDAndSubdirPattern     = regexache.MustCompile(`^([0-9A-Za-z.-]+)(?::\d{0,5})?(/.*)$`)
	s3OutpostsAccessPointARNResourcePattern = regexache.MustCompile(`^outpost/.*/accesspoint/.*?(/.*)$`)
)

// globalIDFromLocationURI extracts the global ID from a location URI.
// https://docs.aws.amazon.com/datasync/latest/userguide/API_LocationListEntry.html#DataSync-Type-LocationListEntry-LocationUri
func globalIDFromLocationURI(uri string) (string, error) {
	submatches := locationURIPattern.FindStringSubmatch(uri)

	if len(submatches) != 3 {
		return "", fmt.Errorf("location URI (%s) does not match pattern %q", uri, locationURIPattern)
	}

	globalIDAndSubdir := submatches[2]

	submatches = locationURIGlobalIDAndSubdirPattern.FindStringSubmatch(globalIDAndSubdir)

	if len(submatches) != 3 {
		return "", fmt.Errorf("location URI global ID and subdirectory (%s) does not match pattern %q", globalIDAndSubdir, locationURIGlobalIDAndSubdirPattern)
	}

	return submatches[1], nil
}

// subdirectoryFromLocationURI extracts the subdirectory from a location URI.
// https://docs.aws.amazon.com/datasync/latest/userguide/API_LocationListEntry.html#DataSync-Type-LocationListEntry-LocationUri
func subdirectoryFromLocationURI(uri string) (string, error) {
	submatches := locationURIPattern.FindStringSubmatch(uri)

	if len(submatches) != 3 {
		return "", fmt.Errorf("location URI (%s) does not match pattern %q", uri, locationURIPattern)
	}

	globalIDAndSubdir := submatches[2]
	parsedARN, err := arn.Parse(globalIDAndSubdir)

	if err == nil {
		submatches = s3OutpostsAccessPointARNResourcePattern.FindStringSubmatch(parsedARN.Resource)

		if len(submatches) != 2 {
			return "", fmt.Errorf("location URI S3 on Outposts access point ARN resource (%s) does not match pattern %q", parsedARN.Resource, s3OutpostsAccessPointARNResourcePattern)
		}

		return submatches[1], nil
	}

	submatches = locationURIGlobalIDAndSubdirPattern.FindStringSubmatch(globalIDAndSubdir)

	if len(submatches) != 3 {
		return "", fmt.Errorf("location URI global ID and subdirectory (%s) does not match pattern %q", globalIDAndSubdir, locationURIGlobalIDAndSubdirPattern)
	}

	return submatches[2], nil
}
