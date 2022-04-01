package names

import (
	"fmt"
	"os"
	"testing"
)

func TestServiceForHCLKey(t *testing.T) {
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
		t.Run(testCase.TestName, func(t *testing.T) {
			got, err := ServiceForHCLKey(testCase.Input)

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
	nonExisting := []string{
		"alexaforbusiness",
		"amplifybackend",
		"appflow",
		"appintegrations",
		"applicationcostprofiler",
		"applicationdiscovery",
		"applicationinsights",
		"appregistry",
		"auditmanager",
		"augmentedairuntime",
		"braket",
		"clouddirectory",
		"cloudsearchdomain",
		"cloudwatchrum",
		"codeguruprofiler",
		"codegurureviewer",
		"codestar",
		"cognitosync",
		"comprehend",
		"comprehendmedical",
		"connectcontactlens",
		"connectparticipant",
		"costexplorer",
		"devopsguru",
		"dynamodbstreams",
		"ec2instanceconnect",
		"elasticinference",
		"emrcontainers",
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
		"healthlake",
		"honeycode",
		"iot1clickdevices",
		"iot1clickprojects",
		"iotdataplane",
		"iotdeviceadvisor",
		"ioteventsdata",
		"iotfleethub",
		"iotjobsdataplane",
		"iotsecuretunneling",
		"iotsitewise",
		"iotthingsgraph",
		"iotwireless",
		"kendra",
		"kinesisvideoarchivedmedia",
		"kinesisvideomedia",
		"kinesisvideosignalingchannels",
		"lexmodelsv2",
		"lexruntime",
		"lexruntimev2",
		"location",
		"lookoutequipment",
		"lookoutforvision",
		"lookoutmetrics",
		"machinelearning",
		"managedblockchain",
		"marketplacecatalog",
		"marketplacecommerceanalytics",
		"marketplaceentitlement",
		"marketplacemetering",
		"mediapackagevod",
		"mediastoredata",
		"mediatailor",
		"mgn",
		"migrationhub",
		"migrationhubconfig",
		"mobile",
		"mobileanalytics",
		"mturk",
		"nimblestudio",
		"opsworkscm",
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
		"redshiftdata",
		"rekognition",
		"robomaker",
		"sagemakeredgemanager",
		"sagemakerfeaturestoreruntime",
		"sagemakerruntime",
		"savingsplans",
		"sesv2",
		"sms",
		"snowball",
		"ssmcontacts",
		"ssmincidents",
		"sso",
		"ssooidc",
		"support",
		"textract",
		"timestreamquery",
		"transcribe",
		"transcribestreaming",
		"translate",
		"wellarchitected",
		"workdocs",
		"workmail",
		"workmailmessageflow",
	}

	for _, testCase := range ServiceKeys() {
		t.Run(testCase, func(t *testing.T) {
			wd, err := os.Getwd()
			if err != nil {
				t.Errorf("error reading working directory: %s", err)
			}

			if _, err := os.Stat(fmt.Sprintf("%s/../internal/service/%s", wd, testCase)); os.IsNotExist(err) {
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

func TestServiceProviderNameUpper(t *testing.T) {
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
		t.Run(testCase.TestName, func(t *testing.T) {
			got, err := ServiceProviderNameUpper(testCase.Input)

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

func TestAWSServiceName(t *testing.T) {
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
			TestName: AppAutoScaling,
			Input:    AppAutoScaling,
			Expected: "AppAutoScaling",
			Error:    false,
		},
		{
			TestName: DMS,
			Input:    DMS,
			Expected: "DMS",
			Error:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got, err := ServiceProviderNameUpper(testCase.Input)

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
