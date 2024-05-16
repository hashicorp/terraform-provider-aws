// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

func TestNewObjectARN_GeneralPurposeBucket(t *testing.T) {
	t.Parallel()

	expectedARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3",
		Resource:  "test-bucket/test-key",
	}

	arn, err := newObjectARN("test-partition", "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	equalARN(t, arn, expectedARN)
}

func TestNewObjectARN_GeneralPurposeBucket_AccessPointInBucketName(t *testing.T) {
	t.Parallel()

	expectedARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3",
		Resource:  "test-accesspoint-bucket/test-key",
	}

	arn, err := newObjectARN("test-partition", "test-accesspoint-bucket", "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	equalARN(t, arn, expectedARN)
}

func TestNewObjectARN_AccessPoint(t *testing.T) {
	t.Parallel()

	expectedARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3",
		Region:    "us-west-2", //lintignore:AWSAT003
		AccountID: "123456789012",
		Resource:  "accesspoint/test-accesspoint/test-key",
	}

	apARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3",
		Region:    "us-west-2", //lintignore:AWSAT003
		AccountID: "123456789012",
		Resource:  "accesspoint/test-accesspoint",
	}

	arn, err := newObjectARN("ignored", apARN.String(), "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	equalARN(t, arn, expectedARN)
}

func TestNewObjectARN_MultiRegionAccessPoint(t *testing.T) {
	t.Parallel()

	expectedARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3",
		AccountID: "123456789012",
		Resource:  "accesspoint/test-multi-region-accesspoint/test-key",
	}

	mrapARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3",
		AccountID: "123456789012",
		Resource:  "accesspoint/test-multi-region-accesspoint",
	}

	arn, err := newObjectARN("ignored", mrapARN.String(), "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	equalARN(t, arn, expectedARN)
}

func TestNewObjectARN_ObjectLambdaAccessPoint(t *testing.T) {
	t.Parallel()

	expectedARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3-object-lambda",
		AccountID: "123456789012",
		Resource:  "accesspoint/test-object-lambda-accesspoint/test-key",
	}

	olapARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3-object-lambda",
		AccountID: "123456789012",
		Resource:  "accesspoint/test-object-lambda-accesspoint",
	}

	arn, err := newObjectARN("ignored", olapARN.String(), "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	equalARN(t, arn, expectedARN)
}

func TestParseObjectARN_GeneralPurposeBucket(t *testing.T) {
	t.Parallel()

	expectedObjectARN := objectARN{
		ARN: arn.ARN{
			Partition: "test-partition",
			Service:   "s3",
			Resource:  "test-bucket/test-key",
		},
		Bucket: "test-bucket",
		Key:    "test-key",
	}

	oARN, _ := newObjectARN("test-partition", "test-bucket", "test-key")

	parsed, err := parseObjectARN(oARN.String())
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	equalObjectARN(t, parsed, expectedObjectARN)
}

func TestParseObjectARN_GeneralPurposeBucket_AccessPointBucketName(t *testing.T) {
	t.Parallel()

	expectedObjectARN := objectARN{
		ARN: arn.ARN{
			Partition: "test-partition",
			Service:   "s3",
			Resource:  "accesspoint/test-key",
		},
		Bucket: "accesspoint",
		Key:    "test-key",
	}

	oARN, _ := newObjectARN("test-partition", "accesspoint", "test-key")

	parsed, err := parseObjectARN(oARN.String())
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	equalObjectARN(t, parsed, expectedObjectARN)
}

func TestParseObjectARN_AccessPoint(t *testing.T) {
	t.Parallel()

	expectedObjectARN := objectARN{
		ARN: arn.ARN{
			Partition: "test-partition",
			Service:   "s3",
			Region:    "us-west-2", //lintignore:AWSAT003
			AccountID: "123456789012",
			Resource:  "accesspoint/test-accesspoint/test-key",
		},
		Bucket: "arn:test-partition:s3:us-west-2:123456789012:accesspoint/test-accesspoint", //lintignore:AWSAT003
		Key:    "test-key",
	}

	apARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3",
		Region:    "us-west-2", //lintignore:AWSAT003
		AccountID: "123456789012",
		Resource:  "accesspoint/test-accesspoint",
	}

	oARN, _ := newObjectARN("ignored", apARN.String(), "test-key")

	parsed, err := parseObjectARN(oARN.String())
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	equalObjectARN(t, parsed, expectedObjectARN)
}

func TestParseObjectARN_MultiRegionAccessPoint(t *testing.T) {
	t.Parallel()

	expectedObjectARN := objectARN{
		ARN: arn.ARN{
			Partition: "test-partition",
			Service:   "s3",
			AccountID: "123456789012",
			Resource:  "accesspoint/test-multi-region-accesspoint/test-key",
		},
		Bucket: "arn:test-partition:s3::123456789012:accesspoint/test-multi-region-accesspoint",
		Key:    "test-key",
	}

	mrapARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3",
		AccountID: "123456789012",
		Resource:  "accesspoint/test-multi-region-accesspoint",
	}

	oARN, _ := newObjectARN("ignored", mrapARN.String(), "test-key")

	parsed, err := parseObjectARN(oARN.String())
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	equalObjectARN(t, parsed, expectedObjectARN)
}

func TestParseObjectARN_ObjectLambdaAccessPoint(t *testing.T) {
	t.Parallel()

	expectedObjectARN := objectARN{
		ARN: arn.ARN{
			Partition: "test-partition",
			Service:   "s3-object-lambda",
			AccountID: "123456789012",
			Resource:  "accesspoint/test-object-lambda-accesspoint/test-key",
		},
		Bucket: "arn:test-partition:s3-object-lambda::123456789012:accesspoint/test-object-lambda-accesspoint",
		Key:    "test-key",
	}

	olapARN := arn.ARN{
		Partition: "test-partition",
		Service:   "s3-object-lambda",
		AccountID: "123456789012",
		Resource:  "accesspoint/test-object-lambda-accesspoint",
	}

	oARN, _ := newObjectARN("ignored", olapARN.String(), "test-key")

	parsed, err := parseObjectARN(oARN.String())
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	equalObjectARN(t, parsed, expectedObjectARN)
}

func equalARN(t *testing.T, a, e arn.ARN) {
	t.Helper()

	if a.Partition != e.Partition {
		t.Errorf("partition: expected %q, got %q", e.Partition, a.Partition)
	}
	if a.Service != e.Service {
		t.Errorf("service: expected %q, got %q", e.Service, a.Service)
	}
	if a.Region != e.Region {
		t.Errorf("region: expected %q, got %q", e.Region, a.Region)
	}
	if a.AccountID != e.AccountID {
		t.Errorf("account ID: expected %q, got %q", e.AccountID, a.AccountID)
	}
	if a.Resource != e.Resource {
		t.Errorf("resource: expected %q, got %q", e.Resource, a.Resource)
	}
}

func equalObjectARN(t *testing.T, a, e objectARN) {
	t.Helper()

	equalARN(t, a.ARN, e.ARN)
	if a.Bucket != e.Bucket {
		t.Errorf("bucket: expected %q, got %q", e.Bucket, a.Bucket)
	}
	if a.Key != e.Key {
		t.Errorf("key: expected %q, got %q", e.Key, a.Key)
	}
}
