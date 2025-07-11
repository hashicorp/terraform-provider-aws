// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
)

var (
	standardPartition, _ = endpoints.PartitionForRegion(endpoints.DefaultPartitions(), endpoints.UsEast1RegionID)
	chinaPartition, _    = endpoints.PartitionForRegion(endpoints.DefaultPartitions(), endpoints.CnNorth1RegionID)
)

func TestAWSClientPartitionHostname(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx := t.Context()
	testCases := []struct {
		Name      string
		AWSClient *AWSClient
		Prefix    string
		Expected  string
	}{
		{
			Name: "AWS Commercial",
			AWSClient: &AWSClient{
				partition: standardPartition,
			},
			Prefix:   "test",
			Expected: "test.amazonaws.com",
		},
		{
			Name: "AWS China",
			AWSClient: &AWSClient{
				partition: chinaPartition,
			},
			Prefix:   "test",
			Expected: "test.amazonaws.com.cn",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got := testCase.AWSClient.PartitionHostname(ctx, testCase.Prefix)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAWSClientRegionalHostname(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx := t.Context()
	testCases := []struct {
		Name      string
		AWSClient *AWSClient
		Prefix    string
		Expected  string
	}{
		{
			Name: "AWS Commercial",
			AWSClient: &AWSClient{
				partition: standardPartition,
				awsConfig: &aws.Config{
					Region: "us-west-2", //lintignore:AWSAT003
				},
			},
			Prefix:   "test",
			Expected: "test.us-west-2.amazonaws.com", //lintignore:AWSAT003
		},
		{
			Name: "AWS China",
			AWSClient: &AWSClient{
				partition: chinaPartition,
				awsConfig: &aws.Config{
					Region: "cn-northwest-1", //lintignore:AWSAT003
				},
			},
			Prefix:   "test",
			Expected: "test.cn-northwest-1.amazonaws.com.cn", //lintignore:AWSAT003
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got := testCase.AWSClient.RegionalHostname(ctx, testCase.Prefix)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAWSClientEC2PrivateDNSNameForIP(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx := t.Context()
	testCases := []struct {
		Name      string
		AWSClient *AWSClient
		IP        string
		Expected  string
	}{
		{
			Name: "us-west-2",
			AWSClient: &AWSClient{
				partition: standardPartition,
				awsConfig: &aws.Config{
					Region: "us-west-2", //lintignore:AWSAT003
				},
			},
			IP:       "10.20.30.40",
			Expected: "ip-10-20-30-40.us-west-2.compute.internal", //lintignore:AWSAT003
		},
		{
			Name: "us-east-1",
			AWSClient: &AWSClient{
				partition: standardPartition,
				awsConfig: &aws.Config{
					Region: "us-east-1", //lintignore:AWSAT003
				},
			},
			IP:       "10.20.30.40",
			Expected: "ip-10-20-30-40.ec2.internal",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got := testCase.AWSClient.EC2PrivateDNSNameForIP(ctx, testCase.IP)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAWSClientEC2PublicDNSNameForIP(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx := t.Context()
	testCases := []struct {
		Name      string
		AWSClient *AWSClient
		IP        string
		Expected  string
	}{
		{
			Name: "us-west-2",
			AWSClient: &AWSClient{
				partition: standardPartition,
				awsConfig: &aws.Config{
					Region: "us-west-2", //lintignore:AWSAT003
				},
			},
			IP:       "10.20.30.40",
			Expected: "ec2-10-20-30-40.us-west-2.compute.amazonaws.com", //lintignore:AWSAT003
		},
		{
			Name: "us-east-1",
			AWSClient: &AWSClient{
				partition: standardPartition,
				awsConfig: &aws.Config{
					Region: "us-east-1", //lintignore:AWSAT003
				},
			},
			IP:       "10.20.30.40",
			Expected: "ec2-10-20-30-40.compute-1.amazonaws.com",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got := testCase.AWSClient.EC2PublicDNSNameForIP(ctx, testCase.IP)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAWSClientValidateInContextRegionInPartition(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx := t.Context()
	testCases := []struct {
		Name      string
		AWSClient *AWSClient
		Region    string
		Expected  bool
	}{
		{
			Name: "AWS Commercial, valid",
			AWSClient: &AWSClient{
				partition: standardPartition,
			},
			Region:   endpoints.ApNortheast1RegionID,
			Expected: true,
		},
		{
			Name: "AWS Commercial, invalid",
			AWSClient: &AWSClient{
				partition: standardPartition,
			},
			Region:   endpoints.UsGovWest1RegionID,
			Expected: false,
		},
		{
			Name: "AWS China, valid",
			AWSClient: &AWSClient{
				partition: chinaPartition,
			},
			Region:   endpoints.CnNorth1RegionID,
			Expected: true,
		},
		{
			Name:      "Empty partition, valid",
			AWSClient: &AWSClient{},
			Region:    "ash",
			Expected:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			ctx = NewResourceContext(ctx, "test", "Test", testCase.Region)
			err := testCase.AWSClient.ValidateInContextRegionInPartition(ctx)

			if got := err == nil; got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
