// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import "testing"

func TestSubdirectoryFromLocationURI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName             string
		InputURI             string
		ExpectedError        bool
		ExpectedSubdirectory string
	}{
		{
			TestName:      "empty URI",
			InputURI:      "",
			ExpectedError: true,
		},
		{
			TestName:      "invalid URI scheme",
			InputURI:      "test://testing/",
			ExpectedError: true,
		},
		{
			TestName:      "S3 bucket URI no bucket name (1)",
			InputURI:      "s3://",
			ExpectedError: true,
		},
		{
			TestName:      "S3 bucket URI no bucket name (2)",
			InputURI:      "s3:///",
			ExpectedError: true,
		},
		{
			TestName:             "S3 bucket URI top level",
			InputURI:             "s3://bucket/",
			ExpectedSubdirectory: "/",
		},
		{
			TestName:             "S3 bucket URI one level",
			InputURI:             "s3://bucket/my-folder-1/",
			ExpectedSubdirectory: "/my-folder-1/",
		},
		{
			TestName:             "S3 bucket URI two levels",
			InputURI:             "s3://bucket/my-folder-1/my-folder-2",
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		{
			TestName:             "S3 Outposts ARN URI top level",
			InputURI:             "s3://arn:aws:s3-outposts:eu-west-3:123456789012:outpost/op-YYYYYYYYYY/accesspoint/my-access-point/", //lintignore:AWSAT003,AWSAT005
			ExpectedSubdirectory: "/",
		},
		{
			TestName:             "S3 Outposts ARN URI one level",
			InputURI:             "s3://arn:aws:s3-outposts:eu-west-3:123456789012:outpost/op-YYYYYYYYYY/accesspoint/my-access-point/my-folder-1/", //lintignore:AWSAT003,AWSAT005
			ExpectedSubdirectory: "/my-folder-1/",
		},
		{
			TestName:             "S3 Outposts ARN URI two levels",
			InputURI:             "s3://arn:aws:s3-outposts:eu-west-3:123456789012:outpost/op-YYYYYYYYYY/accesspoint/my-access-point/my-folder-1/my-folder-2", //lintignore:AWSAT003,AWSAT005
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		{
			TestName:             "EFS URI top level",
			InputURI:             "efs://us-west-2.fs-abcdef01/", //lintignore:AWSAT003
			ExpectedSubdirectory: "/",
		},
		{
			TestName:             "EFS URI one level",
			InputURI:             "efs://us-west-2.fs-abcdef01/my-folder-1/", //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/",
		},
		{
			TestName:             "EFS URI two levels",
			InputURI:             "efs://us-west-2.fs-abcdef01/my-folder-1/my-folder-2", //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		{
			TestName:             "NFS URI top level",
			InputURI:             "nfs://example.com/",
			ExpectedSubdirectory: "/",
		},
		{
			TestName:             "NFS URI one level",
			InputURI:             "nfs://example.com/my-folder-1/",
			ExpectedSubdirectory: "/my-folder-1/",
		},
		{
			TestName:             "NFS URI two levels",
			InputURI:             "nfs://example.com/my-folder-1/my-folder-2",
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		{
			TestName:             "SMB URI top level",
			InputURI:             "smb://192.168.1.1/",
			ExpectedSubdirectory: "/",
		},
		{
			TestName:             "SMB URI one level",
			InputURI:             "smb://192.168.1.1/my-folder-1/",
			ExpectedSubdirectory: "/my-folder-1/",
		},
		{
			TestName:             "SMB URI two levels",
			InputURI:             "smb://192.168.1.1/my-folder-1/my-folder-2",
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		{
			TestName:             "HDFS URI top level",
			InputURI:             "hdfs://192.168.1.1:80/",
			ExpectedSubdirectory: "/",
		},
		{
			TestName:             "HDFS URI one level",
			InputURI:             "hdfs://192.168.1.1:80/my-folder-1/",
			ExpectedSubdirectory: "/my-folder-1/",
		},
		{
			TestName:             "HDFS URI two levels",
			InputURI:             "hdfs://192.168.1.1:80/my-folder-1/my-folder-2",
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		{
			TestName:             "FSx Windows URI top level",
			InputURI:             "fsxw://us-west-2.fs-abcdef012345678901/", //lintignore:AWSAT003
			ExpectedSubdirectory: "/",
		},
		{
			TestName:             "FSx Windows URI one level",
			InputURI:             "fsxw://us-west-2.fs-abcdef012345678901/my-folder-1/", //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/",
		},
		{
			TestName:             "FSx Windows URI two levels",
			InputURI:             "fsxw://us-west-2.fs-abcdef012345678901/my-folder-1/my-folder-2", //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		{
			TestName:             "FSx Zfs URI top level",
			InputURI:             "fsxz://us-west-2.fs-abcdef012345678901/", //lintignore:AWSAT003
			ExpectedSubdirectory: "/",
		},
		{
			TestName:             "FSx Zfs URI one level",
			InputURI:             "fsxz://us-west-2.fs-abcdef012345678901/my-folder-1/", //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/",
		},
		{
			TestName:             "FSx Zfs URI two levels",
			InputURI:             "fsxz://us-west-2.fs-abcdef012345678901/my-folder-1/my-folder-2", //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		{
			TestName:      "Object storage two levels",
			InputURI:      "object-storage://192.168.1.1/tf-acc-test-5815577519131245007/tf-acc-test-5815577519131245008/",
			ExpectedError: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := subdirectoryFromLocationURI(testCase.InputURI)

			if err == nil && testCase.ExpectedError {
				t.Fatalf("expected error")
			}

			if err != nil && !testCase.ExpectedError {
				t.Fatalf("unexpected error: %s", err)
			}

			if got != testCase.ExpectedSubdirectory {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedSubdirectory)
			}
		})
	}
}

func TestDecodeObjectStorageURI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName             string
		InputURI             string
		ExpectedError        bool
		ExpectedHostname     string
		ExpectedBucketName   string
		ExpectedSubdirectory string
	}{
		{
			TestName:      "empty URI",
			InputURI:      "",
			ExpectedError: true,
		},
		{
			TestName:      "S3 bucket URI top level",
			InputURI:      "s3://bucket/",
			ExpectedError: true,
		},
		{
			TestName:             "Object storage top level",
			InputURI:             "object-storage://tawn19fp.test/tf-acc-test-6405856757419817388/",
			ExpectedHostname:     "tawn19fp.test",
			ExpectedBucketName:   "tf-acc-test-6405856757419817388",
			ExpectedSubdirectory: "/",
		},
		{
			TestName:             "Object storage one level",
			InputURI:             "object-storage://tawn19fp.test/tf-acc-test-6405856757419817388/test",
			ExpectedHostname:     "tawn19fp.test",
			ExpectedBucketName:   "tf-acc-test-6405856757419817388",
			ExpectedSubdirectory: "/test",
		},
		{
			TestName:             "Object storage two levels",
			InputURI:             "object-storage://192.168.1.1/tf-acc-test-5815577519131245007/tf-acc-test-5815577519131245008/tf-acc-test-5815577519131245009/",
			ExpectedHostname:     "192.168.1.1",
			ExpectedBucketName:   "tf-acc-test-5815577519131245007",
			ExpectedSubdirectory: "/tf-acc-test-5815577519131245008/tf-acc-test-5815577519131245009/",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			gotHostname, gotBucketName, gotSubdirectory, err := decodeObjectStorageURI(testCase.InputURI)

			if err == nil && testCase.ExpectedError {
				t.Fatalf("expected error")
			}

			if err != nil && !testCase.ExpectedError {
				t.Fatalf("unexpected error: %s", err)
			}

			if gotHostname != testCase.ExpectedHostname {
				t.Errorf("hostname %s, expected %s", gotHostname, testCase.ExpectedHostname)
			}

			if gotBucketName != testCase.ExpectedBucketName {
				t.Errorf("bucketName %s, expected %s", gotBucketName, testCase.ExpectedBucketName)
			}

			if gotSubdirectory != testCase.ExpectedSubdirectory {
				t.Errorf("subdirectory %s, expected %s", gotSubdirectory, testCase.ExpectedSubdirectory)
			}
		})
	}
}
