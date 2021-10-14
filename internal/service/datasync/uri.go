package datasync

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var (
	locationURIPattern                      = regexp.MustCompile(`^(efs|nfs|s3|smb|fsxw)://(.+)$`)
	locationURIGlobalIDAndSubdirPattern     = regexp.MustCompile(`^([a-zA-Z0-9.\-]+)(/.*)$`)
	s3OutpostsAccessPointARNResourcePattern = regexp.MustCompile(`^outpost/.*/accesspoint/.*?(/.*)$`)
)

// SubdirectoryFromLocationURI extracts the subdirectory from a location URI.
// https://docs.aws.amazon.com/datasync/latest/userguide/API_LocationListEntry.html#DataSync-Type-LocationListEntry-LocationUri
func SubdirectoryFromLocationURI(uri string) (string, error) {
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
