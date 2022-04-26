package names

import (
	"fmt"
	"os"
	"testing"
)

func TestProviderPackageForAlias(t *testing.T) {
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
	nonExisting := []string{
		"alexaforbusiness",
		"amplifybackend",
		"amplifyuibuilder",
		"apigatewaymanagementapi",
		"appconfigdata",
		"appflow",
		"appintegrations",
		"applicationcostprofiler",
		"applicationdiscovery",
		"applicationinsights",
		"appregistry",
		"auditmanager",
		"augmentedairuntime",
		"backupgateway",
		"billingconductor",
		"braket",
		"ce",
		"chimesdkidentity",
		"chimesdkmeetings",
		"chimesdkmessaging",
		"clouddirectory",
		"cloudsearchdomain",
		"cloudwatchevidently",
		"cloudwatchrum",
		"codeguruprofiler",
		"codegurureviewer",
		"codestar",
		"cognitosync",
		"comprehend",
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
		"healthlake",
		"honeycode",
		"inspector2",
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
		"ivs",
		"kendra",
		"kinesisvideoarchivedmedia",
		"kinesisvideomedia",
		"kinesisvideosignaling",
		"kinesisvideosignalingchannels",
		"lexmodelsv2",
		"lexruntime",
		"lexruntimev2",
		"location",
		"lookoutequipment",
		"lookoutforvision",
		"lookoutmetrics",
		"lookoutvision",
		"machinelearning",
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
		"rbin",
		"rdsdata",
		"redshiftdata",
		"rekognition",
		"resiliencehub",
		"robomaker",
		"route53recoverycluster",
		"rum",
		"sagemakera2iruntime",
		"sagemakeredge",
		"sagemakeredgemanager",
		"sagemakerfeaturestoreruntime",
		"sagemakerruntime",
		"savingsplans",
		"servicecatalogappregistry",
		"sesv2",
		"sms",
		"snowball",
		"snowdevicemanagement",
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
		"voiceid",
		"wellarchitected",
		"wisdom",
		"workdocs",
		"workmail",
		"workmailmessageflow",
		"workspacesweb",
	}

	for _, testCase := range ProviderPackages() {
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

func TestProviderNameUpper(t *testing.T) {
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
			TestName: DRS,
			Input:    DRS,
			Expected: "AWS DRS (Elastic Disaster Recovery)",
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

func TestAWSGoV1Package(t *testing.T) {
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
			TestName: "same as AWS",
			Input:    Translate,
			Expected: Translate,
			Error:    false,
		},
		{
			TestName: "different from AWS",
			Input:    Transcribe,
			Expected: "transcribeservice",
			Error:    false,
		},
		{
			TestName: "different from AWS 2",
			Input:    RBin,
			Expected: "recyclebin",
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
			got, err := AWSGoV1Package(testCase.Input)

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
		t.Run(testCase.TestName, func(t *testing.T) {
			got, err := AWSGoV1ClientName(testCase.Input)

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
