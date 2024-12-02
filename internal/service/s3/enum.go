// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

type bucketNameType int

const (
	bucketNameTypeGeneralPurposeBucket bucketNameType = iota
	bucketNameTypeDirectoryBucket
	bucketNameTypeAccessPointAlias
	bucketNameTypeAccessPointARN
	bucketNameTypeObjectLambdaAccessPointAlias
	bucketNameTypeObjectLambdaAccessPointARN
	bucketNameTypeMultiRegionAccessPointARN
)

func bucketNameTypeFor(bucket string) bucketNameType {
	switch {
	case arn.IsARN(bucket):
		v, _ := arn.Parse(bucket)
		switch {
		case strings.HasPrefix(v.Resource, "accesspoint/"):
			switch v.Service {
			case "s3":
				if v.Region == "" {
					return bucketNameTypeMultiRegionAccessPointARN
				}
				return bucketNameTypeAccessPointARN
			case "s3-object-lambda":
				return bucketNameTypeObjectLambdaAccessPointARN
			}
		}
	case directoryBucketNameRegex.MatchString(bucket):
		return bucketNameTypeDirectoryBucket
	case strings.HasSuffix(bucket, "-s3alias"):
		return bucketNameTypeAccessPointAlias
	case strings.HasSuffix(bucket, "--ol-s3"):
		return bucketNameTypeObjectLambdaAccessPointAlias
	}

	return bucketNameTypeGeneralPurposeBucket
}
