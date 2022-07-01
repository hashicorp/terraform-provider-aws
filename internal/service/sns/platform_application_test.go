package sns_test

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsns "github.com/hashicorp/terraform-provider-aws/internal/service/sns"
)

/**
 Before running this test, at least one of these ENV variables combinations must be set:

 GCM_API_KEY - Google Cloud Messaging API Key

 APNS_SANDBOX_CREDENTIAL_PATH - Apple Push Notification Sandbox Private Key file location
 APNS_SANDBOX_PRINCIPAL_PATH - Apple Push Notification Sandbox Certificate file location
**/

type testAccPlatformApplicationPlatform struct {
	Name       string
	Credential string
	Principal  string
}

func testAccPlatformApplicationPlatformFromEnv(t *testing.T) []*testAccPlatformApplicationPlatform {
	platforms := make([]*testAccPlatformApplicationPlatform, 0, 2)

	if os.Getenv("APNS_SANDBOX_CREDENTIAL") != "" {
		if os.Getenv("APNS_SANDBOX_PRINCIPAL") == "" {
			t.Fatalf("APNS_SANDBOX_CREDENTIAL set but missing APNS_SANDBOX_PRINCIPAL")
		}

		platform := &testAccPlatformApplicationPlatform{
			Name:       "APNS_SANDBOX",
			Credential: fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_SANDBOX_CREDENTIAL"))),
			Principal:  fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_SANDBOX_PRINCIPAL"))),
		}
		platforms = append(platforms, platform)
	} else if os.Getenv("APNS_SANDBOX_CREDENTIAL_PATH") != "" {
		if os.Getenv("APNS_SANDBOX_PRINCIPAL_PATH") == "" {
			t.Fatalf("APNS_SANDBOX_CREDENTIAL_PATH set but missing APNS_SANDBOX_PRINCIPAL_PATH")
		}

		platform := &testAccPlatformApplicationPlatform{
			Name:       "APNS_SANDBOX",
			Credential: strconv.Quote(fmt.Sprintf("${file(pathexpand(%q))}", os.Getenv("APNS_SANDBOX_CREDENTIAL_PATH"))),
			Principal:  strconv.Quote(fmt.Sprintf("${file(pathexpand(%q))}", os.Getenv("APNS_SANDBOX_PRINCIPAL_PATH"))),
		}
		platforms = append(platforms, platform)
	}

	if os.Getenv("GCM_API_KEY") != "" {
		platform := &testAccPlatformApplicationPlatform{
			Name:       "GCM",
			Credential: strconv.Quote(os.Getenv("GCM_API_KEY")),
		}
		platforms = append(platforms, platform)
	}

	if len(platforms) == 0 {
		t.Skipf("no SNS Platform Application environment variables found")
	}
	return platforms
}

func TestDecodePlatformApplicationID(t *testing.T) {
	var testCases = []struct {
		Input            string
		ExpectedArn      string
		ExpectedName     string
		ExpectedPlatform string
		ErrCount         int
	}{
		{
			Input:            "arn:aws:sns:us-east-1:123456789012:app/APNS_SANDBOX/myAppName", //lintignore:AWSAT003,AWSAT005
			ExpectedArn:      "arn:aws:sns:us-east-1:123456789012:app/APNS_SANDBOX/myAppName", //lintignore:AWSAT003,AWSAT005
			ExpectedName:     "myAppName",
			ExpectedPlatform: "APNS_SANDBOX",
			ErrCount:         0,
		},
		{
			Input:            "arn:aws:sns:us-east-1:123456789012:app/APNS_SANDBOX/myAppName/extra", //lintignore:AWSAT003,AWSAT005
			ExpectedArn:      "",
			ExpectedName:     "",
			ExpectedPlatform: "",
			ErrCount:         1,
		},
		{
			Input:            "arn:aws:sns:us-east-1:123456789012:endpoint/APNS_SANDBOX/myAppName/someID", //lintignore:AWSAT003,AWSAT005
			ExpectedArn:      "",
			ExpectedName:     "",
			ExpectedPlatform: "",
			ErrCount:         1,
		},
		{
			Input:            "arn:aws:sns:us-east-1:123456789012:APNS_SANDBOX/myAppName", //lintignore:AWSAT003,AWSAT005
			ExpectedArn:      "",
			ExpectedName:     "",
			ExpectedPlatform: "",
			ErrCount:         1,
		},
		{
			Input:            "arn:aws:sns:us-east-1:123456789012:app", //lintignore:AWSAT003,AWSAT005
			ExpectedArn:      "",
			ExpectedName:     "",
			ExpectedPlatform: "",
			ErrCount:         1,
		},
		{
			Input:            "myAppName",
			ExpectedArn:      "",
			ExpectedName:     "",
			ExpectedPlatform: "",
			ErrCount:         1,
		},
	}

	for _, tc := range testCases {
		arn, name, platform, err := tfsns.DecodePlatformApplicationID(tc.Input)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Input, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Input)
		}
		if arn != tc.ExpectedArn {
			t.Fatalf("expected %q to return arn: %s", tc.Input, arn)
		}
		if name != tc.ExpectedName {
			t.Fatalf("expected %q to return name: %s", tc.Input, name)
		}
		if platform != tc.ExpectedPlatform {
			t.Fatalf("expected %q to return platform: %s", tc.Input, platform)
		}
	}
}

func TestAccSNSPlatformApplication_basic(t *testing.T) {
	platforms := testAccPlatformApplicationPlatformFromEnv(t)
	resourceName := "aws_sns_platform_application.test"

	for _, platform := range platforms {
		name := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())
		platformPrincipalCheck := resource.TestCheckNoResourceAttr(resourceName, "platform_principal")
		if platform.Principal != "" {
			platformPrincipalCheck = resource.TestCheckResourceAttrSet(resourceName, "platform_principal")
		}

		t.Run(platform.Name, func(*testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:          func() { acctest.PreCheck(t) },
				ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
				ProviderFactories: acctest.ProviderFactories,
				CheckDestroy:      testAccCheckPlatformApplicationDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccPlatformApplicationConfig_basic(name, platform),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckPlatformApplicationExists(resourceName),
							acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sns", regexp.MustCompile(fmt.Sprintf("app/%s/%s$", platform.Name, name))),
							resource.TestCheckResourceAttr(resourceName, "name", name),
							resource.TestCheckResourceAttr(resourceName, "platform", platform.Name),
							resource.TestCheckResourceAttrSet(resourceName, "platform_credential"),
							platformPrincipalCheck,
						),
					},
					{
						ResourceName:            resourceName,
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{"platform_credential", "platform_principal"},
					},
				},
			})
		})
	}
}

func TestAccSNSPlatformApplication_basicAttributes(t *testing.T) {
	platforms := testAccPlatformApplicationPlatformFromEnv(t)
	resourceName := "aws_sns_platform_application.test"

	var testCases = []struct {
		AttributeKey         string
		AttributeValue       string
		AttributeValueUpdate string
	}{
		{
			AttributeKey:         "success_feedback_sample_rate",
			AttributeValue:       "100",
			AttributeValueUpdate: "99",
		},
	}

	for _, platform := range platforms {
		t.Run(platform.Name, func(*testing.T) {
			for _, tc := range testCases {
				t.Run(fmt.Sprintf("%s/%s", platform.Name, tc.AttributeKey), func(*testing.T) {
					name := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())

					resource.ParallelTest(t, resource.TestCase{
						PreCheck:          func() { acctest.PreCheck(t) },
						ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
						ProviderFactories: acctest.ProviderFactories,
						CheckDestroy:      testAccCheckPlatformApplicationDestroy,
						Steps: []resource.TestStep{
							{
								Config: testAccPlatformApplicationConfig_basicAttribute(name, platform, tc.AttributeKey, tc.AttributeValue),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckPlatformApplicationExists(resourceName),
									resource.TestCheckResourceAttr(resourceName, tc.AttributeKey, tc.AttributeValue),
								),
							},
							{
								Config: testAccPlatformApplicationConfig_basicAttribute(name, platform, tc.AttributeKey, tc.AttributeValueUpdate),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckPlatformApplicationExists(resourceName),
									resource.TestCheckResourceAttr(resourceName, tc.AttributeKey, tc.AttributeValueUpdate),
								),
							},
							{
								ResourceName:            resourceName,
								ImportState:             true,
								ImportStateVerify:       true,
								ImportStateVerifyIgnore: []string{"platform_credential", "platform_principal"},
							},
						},
					})
				})
			}
		})
	}
}

func TestAccSNSPlatformApplication_iamRoleAttributes(t *testing.T) {
	platforms := testAccPlatformApplicationPlatformFromEnv(t)
	resourceName := "aws_sns_platform_application.test"

	var testCases = []string{
		"failure_feedback_role_arn",
		"success_feedback_role_arn",
	}

	for _, platform := range platforms {
		t.Run(platform.Name, func(*testing.T) {
			for _, tc := range testCases {
				t.Run(fmt.Sprintf("%s/%s", platform.Name, tc), func(*testing.T) {
					iamRoleName1 := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())
					iamRoleName2 := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())
					name := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())

					resource.ParallelTest(t, resource.TestCase{
						PreCheck:          func() { acctest.PreCheck(t) },
						ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
						ProviderFactories: acctest.ProviderFactories,
						CheckDestroy:      testAccCheckPlatformApplicationDestroy,
						Steps: []resource.TestStep{
							{
								Config: testAccPlatformApplicationConfig_iamRoleAttribute(name, platform, tc, iamRoleName1),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckPlatformApplicationExists(resourceName),
									resource.TestMatchResourceAttr(resourceName, tc, regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:iam::[^:]+:role/%s$", iamRoleName1))),
								),
							},
							{
								Config: testAccPlatformApplicationConfig_iamRoleAttribute(name, platform, tc, iamRoleName2),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckPlatformApplicationExists(resourceName),
									resource.TestMatchResourceAttr(resourceName, tc, regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:iam::[^:]+:role/%s$", iamRoleName2))),
								),
							},
							{
								ResourceName:            resourceName,
								ImportState:             true,
								ImportStateVerify:       true,
								ImportStateVerifyIgnore: []string{"platform_credential", "platform_principal"},
							},
						},
					})
				})
			}
		})
	}
}

func TestAccSNSPlatformApplication_snsTopicAttributes(t *testing.T) {
	platforms := testAccPlatformApplicationPlatformFromEnv(t)
	resourceName := "aws_sns_platform_application.test"

	var testCases = []string{
		"event_delivery_failure_topic_arn",
		"event_endpoint_created_topic_arn",
		"event_endpoint_deleted_topic_arn",
		"event_endpoint_updated_topic_arn",
	}

	for _, platform := range platforms {
		t.Run(platform.Name, func(*testing.T) {
			for _, tc := range testCases {
				t.Run(fmt.Sprintf("%s/%s", platform.Name, tc), func(*testing.T) {
					snsTopicName1 := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())
					snsTopicName2 := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())
					name := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())

					resource.ParallelTest(t, resource.TestCase{
						PreCheck:          func() { acctest.PreCheck(t) },
						ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
						ProviderFactories: acctest.ProviderFactories,
						CheckDestroy:      testAccCheckPlatformApplicationDestroy,
						Steps: []resource.TestStep{
							{
								Config: testAccPlatformApplicationConfig_topicAttribute(name, platform, tc, snsTopicName1),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckPlatformApplicationExists(resourceName),
									resource.TestMatchResourceAttr(resourceName, tc, regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:sns:[^:]+:[^:]+:%s$", snsTopicName1))),
								),
							},
							{
								Config: testAccPlatformApplicationConfig_topicAttribute(name, platform, tc, snsTopicName2),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckPlatformApplicationExists(resourceName),
									resource.TestMatchResourceAttr(resourceName, tc, regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:sns:[^:]+:[^:]+:%s$", snsTopicName2))),
								),
							},
							{
								ResourceName:            resourceName,
								ImportState:             true,
								ImportStateVerify:       true,
								ImportStateVerifyIgnore: []string{"platform_credential", "platform_principal"},
							},
						},
					})
				})
			}
		})
	}
}

func testAccCheckPlatformApplicationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("missing ID: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSConn

		input := &sns.GetPlatformApplicationAttributesInput{
			PlatformApplicationArn: aws.String(rs.Primary.ID),
		}

		log.Printf("[DEBUG] Reading SNS Platform Application attributes: %s", input)
		_, err := conn.GetPlatformApplicationAttributes(input)

		return err
	}
}

func testAccCheckPlatformApplicationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SNSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sns_platform_application" {
			continue
		}

		input := &sns.GetPlatformApplicationAttributesInput{
			PlatformApplicationArn: aws.String(rs.Primary.ID),
		}

		log.Printf("[DEBUG] Reading SNS Platform Application attributes: %s", input)
		_, err := conn.GetPlatformApplicationAttributes(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, sns.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccPlatformApplicationConfig_basic(name string, platform *testAccPlatformApplicationPlatform) string {
	if platform.Principal == "" {
		return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = "%s"
  platform            = "%s"
  platform_credential = %s
}
`, name, platform.Name, platform.Credential)
	}
	return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = "%s"
  platform            = "%s"
  platform_credential = %s
  platform_principal  = %s
}
`, name, platform.Name, platform.Credential, platform.Principal)
}

func testAccPlatformApplicationConfig_basicAttribute(name string, platform *testAccPlatformApplicationPlatform, attributeKey, attributeValue string) string {
	if platform.Principal == "" {
		return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = "%s"
  platform            = "%s"
  platform_credential = %s
  %s                  = "%s"
}
`, name, platform.Name, platform.Credential, attributeKey, attributeValue)
	}
	return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = "%s"
  platform            = "%s"
  platform_credential = %s
  platform_principal  = %s
  %s                  = "%s"
}
`, name, platform.Name, platform.Credential, platform.Principal, attributeKey, attributeValue)
}

func testAccPlatformApplicationConfig_iamRoleAttribute(name string, platform *testAccPlatformApplicationPlatform, attributeKey, iamRoleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "sns.${data.aws_partition.current.dns_suffix}"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

  name = "%s"
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/CloudWatchLogsFullAccess"
  role       = aws_iam_role.test.id
}

%s
`, iamRoleName, testAccPlatformApplicationConfig_basicAttribute(name, platform, attributeKey, "${aws_iam_role.test.arn}"))
}

func testAccPlatformApplicationConfig_topicAttribute(name string, platform *testAccPlatformApplicationPlatform, attributeKey, snsTopicName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = "%s"
}

%s
`, snsTopicName, testAccPlatformApplicationConfig_basicAttribute(name, platform, attributeKey, "${aws_sns_topic.test.arn}"))
}
