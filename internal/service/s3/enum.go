// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

const defaultKMSKeyAlias = "alias/aws/s3"

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
		switch v, _ := arn.Parse(bucket); v.Resource {
		case "accesspoint":
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

	return bucketNameTypeDirectoryBucket
}
