// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

func newObjectARN(partition string, bucket, key string) (arn.ARN, error) {
	if arn.IsARN(bucket) {
		bucketARN, err := arn.Parse(bucket)
		if err != nil {
			return arn.ARN{}, fmt.Errorf("S3 Object ARN: unexpected bucket ARN: %s", bucket)
		}
		bucketARN.Resource = fmt.Sprintf("%s/%s", bucketARN.Resource, key)
		return bucketARN, nil
	}
	return arn.ARN{
		Partition: partition,
		Service:   "s3",
		Resource:  fmt.Sprintf("%s/%s", bucket, key),
	}, nil
}

type objectARN struct {
	arn.ARN
	Bucket string
	Key    string
}

func parseObjectARN(s string) (objectARN, error) {
	arn, err := arn.Parse(s)
	if err != nil {
		return objectARN{}, err
	}

	result := objectARN{
		ARN: arn,
	}

	if strings.HasPrefix(arn.Resource, "accesspoint/") {
		re := regexache.MustCompile(`^(arn:[^:]+:[^:]+:[^:]*:[^:]*:accesspoint/[^/]+)/(.+)$`)
		m := re.FindStringSubmatch(s)
		if len(m) == 3 {
			result.Bucket = m[1]
			result.Key = m[2]
			return result, nil
		}
	}

	parts := strings.SplitN(arn.Resource, "/", 2)
	if len(parts) != 2 {
		return objectARN{}, fmt.Errorf("S3 Object ARN: unexpected resource section: %s", arn.Resource)
	}
	result.Bucket = parts[0]
	result.Key = parts[1]

	return result, nil
}
