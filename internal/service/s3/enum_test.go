// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"testing"
)

func TestBucketNameTypeFor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName     string
		bucket       string
		expectedType bucketNameType
	}{
		{
			testName:     "General purpose bucket name",
			bucket:       "tf-acc-test-5488849387206835662",
			expectedType: bucketNameTypeGeneralPurposeBucket,
		},
		{
			testName:     "Directory bucket name (AZ)",
			bucket:       "tf-acc-test-5488849387206835662--use1-az6--x-s3",
			expectedType: bucketNameTypeDirectoryBucket,
		},
		{
			testName:     "Directory bucket name (medium DLZ)",
			bucket:       "mybucket--test1-zone-ab1--x-s3",
			expectedType: bucketNameTypeDirectoryBucket,
		},
		{
			testName:     "Directory bucket name (long DLZ)",
			bucket:       "mybucket--test1-long1-zone-ab1--x-s3",
			expectedType: bucketNameTypeDirectoryBucket,
		},
		{
			testName:     "Multi-Region access point ARN",
			bucket:       "arn:aws:s3::111122223333:accesspoint/MultiRegionAccessPoint_alias", //lintignore:AWSAT003,AWSAT005
			expectedType: bucketNameTypeMultiRegionAccessPointARN,
		},
		{
			testName:     "Access point ARN",
			bucket:       "arn:aws:s3:us-east-1:111122223333:accesspoint/my-access-point", //lintignore:AWSAT003,AWSAT005
			expectedType: bucketNameTypeAccessPointARN,
		},
		{
			testName:     "Access point alias",
			bucket:       "my-access-point-hrzrlukc5m36ft7okagglf3gmwluquse1b-s3alias",
			expectedType: bucketNameTypeAccessPointAlias,
		},
		{
			testName:     "Object lambda access point alias",
			bucket:       "my-object-lambda-acc-1a4n8yjrb3kda96f67zwrwiiuse1a--ol-s3",
			expectedType: bucketNameTypeObjectLambdaAccessPointAlias,
		},
		{
			testName:     "Object lambda access point ARN",
			bucket:       "arn:aws:s3-object-lambda:us-east-1:111122223333:accesspoint/my-object-lambda-access-point", //lintignore:AWSAT003,AWSAT005
			expectedType: bucketNameTypeObjectLambdaAccessPointARN,
		},
		{
			testName:     "S3 on Outposts access point alias",
			bucket:       "my-access-po-o01ac5d28a6a232904e8xz5w8ijx1qzlbp3i3kuse10--op-s3",
			expectedType: bucketNameTypeS3OnOutpostsAccessPointAlias,
		},
		{
			testName:     "S3 on Outposts access point ARN",
			bucket:       "arn:aws:s3-outposts:us-east-1:123456789012:outpost/op-01ac5d28a6a232904/accesspoint/my-access-point", //lintignore:AWSAT003,AWSAT005
			expectedType: bucketNameTypeS3OnOutpostsAccessPointARN,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			if got, want := bucketNameTypeFor(testCase.bucket), testCase.expectedType; got != want {
				t.Errorf("bucketNameTypeFor(%q) = %v, want %v", testCase.bucket, got, want)
			}
		})
	}
}
