// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
)

func TestARNForNewRegion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		ARN           string
		NewRegion     string
		ExpectedARN   string
		ErrorExpected bool
	}{
		{
			TestName:      acctest.CtBasic,
			ARN:           "arn:aws:dynamodb:us-west-2:786648903940:table/tf-acc-test-7864711876941043153", //lintignore:AWSAT003,AWSAT005
			NewRegion:     "us-east-2",                                                                     //lintignore:AWSAT003
			ExpectedARN:   "arn:aws:dynamodb:us-east-2:786648903940:table/tf-acc-test-7864711876941043153", //lintignore:AWSAT003,AWSAT005
			ErrorExpected: false,
		},
		{
			TestName:      "basic2",
			ARN:           "arn:aws:dynamodb:us-west-2:786648903940:table/tf-acc-test-7864711876941043153", //lintignore:AWSAT003,AWSAT005
			NewRegion:     "us-east-1",                                                                     //lintignore:AWSAT003
			ExpectedARN:   "arn:aws:dynamodb:us-east-1:786648903940:table/tf-acc-test-7864711876941043153", //lintignore:AWSAT003,AWSAT005
			ErrorExpected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := tfdynamodb.ARNForNewRegion(testCase.ARN, testCase.NewRegion)

			if err != nil && !testCase.ErrorExpected {
				t.Errorf("did not expect an error but got one: %s", err)
			}

			if err == nil && testCase.ErrorExpected {
				t.Error("expected an error but got none")
			}

			if got != testCase.ExpectedARN {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedARN)
			}
		})
	}
}

func TestSameAccountReplicaDescriptions(t *testing.T) {
	t.Parallel()

	tableARN := "arn:aws:dynamodb:us-west-2:111122223333:table/test" //lintignore:AWSAT003,AWSAT005

	testCases := []struct {
		TestName        string
		TableARN        string
		Replicas        []awstypes.ReplicaDescription
		ExpectedRegions []string
	}{
		{
			TestName:        "no replicas",
			TableARN:        tableARN,
			ExpectedRegions: []string{},
		},
		{
			TestName: "replicas without ARN are kept",
			TableARN: tableARN,
			Replicas: []awstypes.ReplicaDescription{
				{RegionName: aws.String("us-east-1")}, //lintignore:AWSAT003
			},
			ExpectedRegions: []string{"us-east-1"}, //lintignore:AWSAT003
		},
		{
			TestName: "other account replicas are dropped",
			TableARN: tableARN,
			Replicas: []awstypes.ReplicaDescription{
				{
					RegionName: aws.String("us-east-1"),                                          //lintignore:AWSAT003
					ReplicaArn: aws.String("arn:aws:dynamodb:us-east-1:111122223333:table/test"), //lintignore:AWSAT003,AWSAT005
				},
				{
					RegionName: aws.String("us-east-2"),                                          //lintignore:AWSAT003
					ReplicaArn: aws.String("arn:aws:dynamodb:us-east-2:444455556666:table/test"), //lintignore:AWSAT003,AWSAT005
				},
			},
			ExpectedRegions: []string{"us-east-1"}, //lintignore:AWSAT003
		},
		{
			TestName: "unparseable table ARN keeps all replicas",
			TableARN: "not-an-arn",
			Replicas: []awstypes.ReplicaDescription{
				{
					RegionName: aws.String("us-east-2"),                                          //lintignore:AWSAT003
					ReplicaArn: aws.String("arn:aws:dynamodb:us-east-2:444455556666:table/test"), //lintignore:AWSAT003,AWSAT005
				},
			},
			ExpectedRegions: []string{"us-east-2"}, //lintignore:AWSAT003
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got := []string{}
			for _, v := range tfdynamodb.SameAccountReplicaDescriptions(testCase.TableARN, testCase.Replicas) {
				got = append(got, aws.ToString(v.RegionName))
			}

			if diff := cmp.Diff(testCase.ExpectedRegions, got); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
