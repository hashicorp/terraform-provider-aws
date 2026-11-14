// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestParseLocationURI(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		InputURI             string
		ExpectError          bool
		ExpectedGlobalID     string
		ExpectedSubdirectory string
	}{
		"empty URI": {
			InputURI:    "",
			ExpectError: true,
		},
		"invalid URI scheme": {
			InputURI:    "test://testing/",
			ExpectError: true,
		},
		"S3 bucket URI no bucket name (1)": {
			InputURI:    "s3://",
			ExpectError: true,
		},
		"S3 bucket URI no bucket name (2)": {
			InputURI:    "s3:///",
			ExpectError: true,
		},
		"S3 bucket URI top level": {
			InputURI:             "s3://bucket/",
			ExpectedGlobalID:     names.AttrBucket,
			ExpectedSubdirectory: "/",
		},
		"S3 bucket URI one level": {
			InputURI:             "s3://bucket/my-folder-1/",
			ExpectedGlobalID:     names.AttrBucket,
			ExpectedSubdirectory: "/my-folder-1/",
		},
		"S3 bucket URI two levels": {
			InputURI:             "s3://bucket/my-folder-1/my-folder-2",
			ExpectedGlobalID:     names.AttrBucket,
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		"S3 Outposts ARN URI top level": {
			InputURI:             "s3://arn:aws:s3-outposts:eu-west-3:123456789012:outpost/op-YYYYYYYYYY/accesspoint/my-access-point/", //lintignore:AWSAT003,AWSAT005
			ExpectedGlobalID:     "arn:aws:s3-outposts:eu-west-3:123456789012:outpost/op-YYYYYYYYYY/accesspoint/my-access-point",       //lintignore:AWSAT003,AWSAT005
			ExpectedSubdirectory: "/",
		},
		"S3 Outposts ARN URI one level": {
			InputURI:             "s3://arn:aws:s3-outposts:eu-west-3:123456789012:outpost/op-YYYYYYYYYY/accesspoint/my-access-point/my-folder-1/", //lintignore:AWSAT003,AWSAT005
			ExpectedGlobalID:     "arn:aws:s3-outposts:eu-west-3:123456789012:outpost/op-YYYYYYYYYY/accesspoint/my-access-point",                   //lintignore:AWSAT003,AWSAT005
			ExpectedSubdirectory: "/my-folder-1/",
		},
		"S3 Outposts ARN URI two levels": {
			InputURI:             "s3://arn:aws:s3-outposts:eu-west-3:123456789012:outpost/op-YYYYYYYYYY/accesspoint/my-access-point/my-folder-1/my-folder-2", //lintignore:AWSAT003,AWSAT005
			ExpectedGlobalID:     "arn:aws:s3-outposts:eu-west-3:123456789012:outpost/op-YYYYYYYYYY/accesspoint/my-access-point",                              //lintignore:AWSAT003,AWSAT005
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		"EFS URI top level": {
			InputURI:             "efs://us-west-2.fs-abcdef01/", //lintignore:AWSAT003
			ExpectedGlobalID:     "us-west-2.fs-abcdef01",        //lintignore:AWSAT003
			ExpectedSubdirectory: "/",
		},
		"EFS URI one level": {
			InputURI:             "efs://us-west-2.fs-abcdef01/my-folder-1/", //lintignore:AWSAT003
			ExpectedGlobalID:     "us-west-2.fs-abcdef01",                    //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/",
		},
		"EFS URI two levels": {
			InputURI:             "efs://us-west-2.fs-abcdef01/my-folder-1/my-folder-2", //lintignore:AWSAT003
			ExpectedGlobalID:     "us-west-2.fs-abcdef01",                               //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		"NFS URI top level": {
			InputURI:             "nfs://example.com/",
			ExpectedGlobalID:     "example.com",
			ExpectedSubdirectory: "/",
		},
		"NFS URI one level": {
			InputURI:             "nfs://example.com/my-folder-1/",
			ExpectedGlobalID:     "example.com",
			ExpectedSubdirectory: "/my-folder-1/",
		},
		"NFS URI two levels": {
			InputURI:             "nfs://example.com/my-folder-1/my-folder-2",
			ExpectedGlobalID:     "example.com",
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		"SMB URI top level": {
			InputURI:             "smb://192.168.1.1/",
			ExpectedGlobalID:     "192.168.1.1",
			ExpectedSubdirectory: "/",
		},
		"SMB URI one level": {
			InputURI:             "smb://192.168.1.1/my-folder-1/",
			ExpectedGlobalID:     "192.168.1.1",
			ExpectedSubdirectory: "/my-folder-1/",
		},
		"SMB URI two levels": {
			InputURI:             "smb://192.168.1.1/my-folder-1/my-folder-2",
			ExpectedGlobalID:     "192.168.1.1",
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		"HDFS URI top level": {
			InputURI:             "hdfs://192.168.1.1:80/",
			ExpectedGlobalID:     "192.168.1.1",
			ExpectedSubdirectory: "/",
		},
		"HDFS URI one level": {
			InputURI:             "hdfs://192.168.1.1:80/my-folder-1/",
			ExpectedGlobalID:     "192.168.1.1",
			ExpectedSubdirectory: "/my-folder-1/",
		},
		"HDFS URI two levels": {
			InputURI:             "hdfs://192.168.1.1:80/my-folder-1/my-folder-2",
			ExpectedGlobalID:     "192.168.1.1",
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		"FSx Windows URI top level": {
			InputURI:             "fsxw://us-west-2.fs-abcdef123456789012/", //lintignore:AWSAT003
			ExpectedGlobalID:     "us-west-2.fs-abcdef123456789012",         //lintignore:AWSAT003
			ExpectedSubdirectory: "/",
		},
		"FSx Windows URI one level": {
			InputURI:             "fsxw://us-west-2.fs-abcdef123456789012/my-folder-1/", //lintignore:AWSAT003
			ExpectedGlobalID:     "us-west-2.fs-abcdef123456789012",                     //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/",
		},
		"FSx Windows URI two levels": {
			InputURI:             "fsxw://us-west-2.fs-abcdef123456789012/my-folder-1/my-folder-2", //lintignore:AWSAT003
			ExpectedGlobalID:     "us-west-2.fs-abcdef123456789012",                                //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		"FSx Zfs URI top level": {
			InputURI:             "fsxz://us-west-2.fs-abcdef123456789012/", //lintignore:AWSAT003
			ExpectedGlobalID:     "us-west-2.fs-abcdef123456789012",         //lintignore:AWSAT003
			ExpectedSubdirectory: "/",
		},
		"FSx Zfs URI one level": {
			InputURI:             "fsxz://us-west-2.fs-abcdef123456789012/my-folder-1/", //lintignore:AWSAT003
			ExpectedGlobalID:     "us-west-2.fs-abcdef123456789012",                     //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/",
		},
		"FSx Zfs URI two levels": {
			InputURI:             "fsxz://us-west-2.fs-abcdef123456789012/my-folder-1/my-folder-2", //lintignore:AWSAT003
			ExpectedGlobalID:     "us-west-2.fs-abcdef123456789012",                                //lintignore:AWSAT003
			ExpectedSubdirectory: "/my-folder-1/my-folder-2",
		},
		"Object storage two levels": {
			InputURI:    "object-storage://192.168.1.1/tf-acc-test-5815577519131245007/tf-acc-test-5815577519131245008/",
			ExpectError: true,
		},
		"Azure blob URI one level": {
			InputURI:             "azure-blob://example.com/path/",
			ExpectedGlobalID:     "example.com",
			ExpectedSubdirectory: "/path/",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			globalID, subdirectory, err := parseLocationURI(testCase.InputURI)

			if err == nil && testCase.ExpectError {
				t.Fatalf("expected error")
			}

			if err != nil && !testCase.ExpectError {
				t.Fatalf("unexpected error: %s", err)
			}

			if globalID != testCase.ExpectedGlobalID {
				t.Errorf("global ID: got %s, expected %s", globalID, testCase.ExpectedGlobalID)
			}
			if subdirectory != testCase.ExpectedSubdirectory {
				t.Errorf("subdirectory: got %s, expected %s", subdirectory, testCase.ExpectedSubdirectory)
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
