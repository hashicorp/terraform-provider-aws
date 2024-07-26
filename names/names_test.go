// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package names

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"testing"
)

func TestDNSSuffixForPartition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "China",
			input:    ChinaPartitionID,
			expected: "amazonaws.com.cn",
		},
		{
			name:     "GovCloud",
			input:    USGovCloudPartitionID,
			expected: "amazonaws.com",
		},
		{
			name:     "standard",
			input:    StandardPartitionID,
			expected: "amazonaws.com",
		},
		{
			name:     "default",
			input:    "custom",
			expected: "amazonaws.com",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got, want := DNSSuffixForPartition(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %s, expected: %s", got, want)
			}
		})
	}
}

func TestIsOptInRegion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty",
			input:    "",
			expected: false,
		},
		{
			name:     "China",
			input:    CNNorth1RegionID,
			expected: false,
		},
		{
			name:     "GovCloud",
			input:    USGovWest1RegionID,
			expected: false,
		},
		{
			name:     "standard opt-in",
			input:    CAWest1RegionID,
			expected: true,
		},
		{
			name:     "standard not opt-in",
			input:    CACentral1RegionID,
			expected: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got, want := IsOptInRegion(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %t, expected: %t", got, want)
			}
		})
	}
}

func TestPartitionForRegion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "China",
			input:    CNNorth1RegionID,
			expected: ChinaPartitionID,
		},
		{
			name:     "GovCloud",
			input:    USGovWest1RegionID,
			expected: USGovCloudPartitionID,
		},
		{
			name:     "standard",
			input:    USWest2RegionID,
			expected: StandardPartitionID,
		},
		{
			name:     "default",
			input:    "custom",
			expected: StandardPartitionID,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got, want := PartitionForRegion(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %s, expected: %s", got, want)
			}
		})
	}
}

func TestReverseDNS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "amazonaws.com",
			input:    "amazonaws.com",
			expected: "com.amazonaws",
		},
		{
			name:     "amazonaws.com.cn",
			input:    "amazonaws.com.cn",
			expected: "cn.com.amazonaws",
		},
		{
			name:     "sc2s.sgov.gov",
			input:    "sc2s.sgov.gov",
			expected: "gov.sgov.sc2s",
		},
		{
			name:     "c2s.ic.gov",
			input:    "c2s.ic.gov",
			expected: "gov.ic.c2s",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got, want := ReverseDNS(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %s, expected: %s", got, want)
			}
		})
	}
}

func TestProviderPackageForAlias(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected string
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
			Error:    true,
		},
		{
			TestName: "alternate",
			Input:    "transcribeservice",
			Expected: Transcribe,
			Error:    false,
		},
		{
			TestName: "primary",
			Input:    "cognitoidp",
			Expected: CognitoIDP,
			Error:    false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := ProviderPackageForAlias(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestServicesForDirectories(t *testing.T) {
	t.Parallel()

	nonExisting := []string{
		"alexaforbusiness",
		"amplifybackend",
		"amplifyuibuilder",
		"apigatewaymanagementapi",
		"appconfigdata",
		"applicationcostprofiler",
		"applicationdiscovery",
		"applicationinsights",
		"appregistry",
		"augmentedairuntime",
		"backupgateway",
		"billingconductor",
		"braket",
		"chimesdkidentity",
		"chimesdkmeetings",
		"chimesdkmessaging",
		"clouddirectory",
		"cloudsearchdomain",
		"codeguruprofiler",
		"codestar",
		"cognitosync",
		"comprehendmedical",
		"computeoptimizer",
		"connectcontactlens",
		"connectparticipant",
		"costexplorer",
		"customerprofiles",
		"databrew",
		"devopsguru",
		"discovery",
		"drs",
		"dynamodbstreams",
		"ebs",
		"ec2instanceconnect",
		"elasticinference",
		"emrcontainers",
		"evidently",
		"finspace",
		"finspacedata",
		"fis",
		"forecast",
		"forecastquery",
		"frauddetector",
		"gluedatabrew",
		"greengrassv2",
		"groundstation",
		"health",
		"honeycode",
		"iot1clickdevices",
		"iot1clickprojects",
		"iotdata",
		"iotdataplane",
		"iotdeviceadvisor",
		"ioteventsdata",
		"iotfleethub",
		"iotjobsdata",
		"iotjobsdataplane",
		"iotsecuretunneling",
		"iotsitewise",
		"iotthingsgraph",
		"iottwinmaker",
		"iotwireless",
		"kinesisvideoarchivedmedia",
		"kinesisvideomedia",
		"kinesisvideosignaling",
		"kinesisvideosignalingchannels",
		"lexmodelsv2",
		"lexruntime",
		"lexruntimev2",
		"lookoutequipment",
		"lookoutforvision",
		"lookoutmetrics",
		"lookoutvision",
		"machinelearning",
		"macie",
		"managedblockchain",
		"marketplacecatalog",
		"marketplacecommerceanalytics",
		"marketplaceentitlement",
		"marketplacemetering",
		"mediapackagevod",
		"mediastoredata",
		"mediatailor",
		"mgh",
		"mgn",
		"migrationhub",
		"migrationhubconfig",
		"migrationhubrefactorspaces",
		"migrationhubstrategy",
		"mobile",
		"mobileanalytics",
		"mturk",
		"nimble",
		"nimblestudio",
		"opsworkscm",
		"panorama",
		"personalize",
		"personalizeevents",
		"personalizeruntime",
		"pi",
		"pinpointemail",
		"pinpointsmsvoice",
		"polly",
		"proton",
		"qldbsession",
		"rdsdata",
		"rekognition",
		"resiliencehub",
		"robomaker",
		"route53recoverycluster",
		"sagemakera2iruntime",
		"sagemakeredge",
		"sagemakeredgemanager",
		"sagemakerfeaturestoreruntime",
		"sagemakerruntime",
		"savingsplans",
		"servicecatalogappregistry",
		"sms",
		"snowball",
		"snowdevicemanagement",
		"sso",
		"ssooidc",
		"support",
		"textract",
		"timestreamquery",
		"transcribestreaming",
		"translate",
		"voiceid",
		"wellarchitected",
		"wisdom",
		"workdocs",
		"workmail",
		"workmailmessageflow",
		"workspacesweb",
	}

	for _, testCase := range ProviderPackages() {
		testCase := testCase
		t.Run(testCase, func(t *testing.T) {
			t.Parallel()

			wd, err := os.Getwd()
			if err != nil {
				t.Errorf("error reading working directory: %s", err)
			}

			if _, err := os.Stat(fmt.Sprintf("%s/../internal/service/%s", wd, testCase)); errors.Is(err, fs.ErrNotExist) {
				for _, service := range nonExisting {
					if service == testCase {
						t.Skipf("skipping %s because not yet implemented", testCase)
						break
					}
				}
				t.Errorf("expected %s/../internal/service/%s to exist %s", wd, testCase, err)
			}
		})
	}
}

func TestProviderNameUpper(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected string
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
			Error:    true,
		},
		{
			TestName: Transcribe,
			Input:    Transcribe,
			Expected: "Transcribe",
			Error:    false,
		},
		{
			TestName: Route53Domains,
			Input:    Route53Domains,
			Expected: "Route53Domains",
			Error:    false,
		},
		{
			TestName: "doesnotexist",
			Input:    "doesnotexist",
			Expected: "",
			Error:    true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := ProviderNameUpper(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestFullHumanFriendly(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected string
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
			Error:    true,
		},
		{
			TestName: Transcribe,
			Input:    Transcribe,
			Expected: "Amazon Transcribe",
			Error:    false,
		},
		{
			TestName: Synthetics,
			Input:    Synthetics,
			Expected: "Amazon CloudWatch Synthetics",
			Error:    false,
		},
		{
			TestName: "alias",
			Input:    "cloudwatchevidently",
			Expected: "Amazon CloudWatch Evidently",
			Error:    false,
		},
		{
			TestName: "doesnotexist",
			Input:    "doesnotexist",
			Expected: "",
			Error:    true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := FullHumanFriendly(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAWSGoV1ClientName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected string
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
			Error:    true,
		},
		{
			TestName: Elasticsearch,
			Input:    Elasticsearch,
			Expected: "ElasticsearchService",
			Error:    false,
		},
		{
			TestName: Deploy,
			Input:    Deploy,
			Expected: "CodeDeploy",
			Error:    false,
		},
		{
			TestName: RUM,
			Input:    RUM,
			Expected: "CloudWatchRUM",
			Error:    false,
		},
		{
			TestName: CloudControl,
			Input:    CloudControl,
			Expected: "CloudControlApi",
			Error:    false,
		},
		{
			TestName: "doesnotexist",
			Input:    "doesnotexist",
			Expected: "",
			Error:    true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := AWSGoV1ClientTypeName(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}
