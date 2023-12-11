// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsns "github.com/hashicorp/terraform-provider-aws/internal/service/sns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

/**
 Before running this test, at least one of these ENV variables combinations must be set:

 GCM_API_KEY - Google Cloud Messaging API Key

 APNS_SANDBOX_CREDENTIAL - Apple Push Notification Sandbox Private Key
 APNS_SANDBOX_PRINCIPAL - Apple Push Notification Sandbox Certificate

 APNS_SANDBOX_CREDENTIAL_PATH - Apple Push Notification Sandbox Private Key file location
 APNS_SANDBOX_PRINCIPAL_PATH - Apple Push Notification Sandbox Certificate file location

 APNS_SANDBOX_TOKEN_CREDENTIAL - Apple signing key
 APNS_SANDBOX_TOKEN_PRINCIPAL - Apple signing key id
**/

type testAccPlatformApplicationPlatform struct {
	Name         string
	Credential   string
	Principal    string
	ApnsAuthType string // "certificate", "token"
}

func testAccPlatformApplicationPlatformFromEnv(t *testing.T, allowedApnsAuthType string) []*testAccPlatformApplicationPlatform {
	platforms := make([]*testAccPlatformApplicationPlatform, 0, 2)

	if os.Getenv("APNS_SANDBOX_CREDENTIAL") != "" && allowedApnsAuthType == "certificate" {
		if os.Getenv("APNS_SANDBOX_PRINCIPAL") == "" {
			t.Fatalf("APNS_SANDBOX_CREDENTIAL set but missing APNS_SANDBOX_PRINCIPAL")
		}

		platform := &testAccPlatformApplicationPlatform{
			Name:         "APNS_SANDBOX",
			Credential:   fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_SANDBOX_CREDENTIAL"))),
			Principal:    fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_SANDBOX_PRINCIPAL"))),
			ApnsAuthType: "certificate",
		}
		platforms = append(platforms, platform)
	} else if os.Getenv("APNS_SANDBOX_CREDENTIAL_PATH") != "" && allowedApnsAuthType == "certificate" {
		if os.Getenv("APNS_SANDBOX_PRINCIPAL_PATH") == "" {
			t.Fatalf("APNS_SANDBOX_CREDENTIAL_PATH set but missing APNS_SANDBOX_PRINCIPAL_PATH")
		}

		platform := &testAccPlatformApplicationPlatform{
			Name:         "APNS_SANDBOX",
			Credential:   strconv.Quote(fmt.Sprintf("${file(pathexpand(%q))}", os.Getenv("APNS_SANDBOX_CREDENTIAL_PATH"))),
			Principal:    strconv.Quote(fmt.Sprintf("${file(pathexpand(%q))}", os.Getenv("APNS_SANDBOX_PRINCIPAL_PATH"))),
			ApnsAuthType: "certificate",
		}
		platforms = append(platforms, platform)
	} else if os.Getenv("APNS_SANDBOX_TOKEN_CREDENTIAL") != "" && allowedApnsAuthType == "token" {
		if os.Getenv("APNS_SANDBOX_TOKEN_PRINCIPAL") == "" {
			t.Fatalf("APNS_SANDBOX_TOKEN_PRINCIPAL set but missing APNS_SANDBOX_TOKEN_CREDENTIAL")
		}

		platform := &testAccPlatformApplicationPlatform{
			Name:         "APNS_SANDBOX",
			Credential:   fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_SANDBOX_TOKEN_CREDENTIAL"))),
			Principal:    fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_SANDBOX_TOKEN_PRINCIPAL"))),
			ApnsAuthType: "token",
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

func TestParsePlatformApplicationResourceID(t *testing.T) {
	t.Parallel()

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
		arn, name, platform, err := tfsns.ParsePlatformApplicationResourceID(tc.Input)
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

func TestAccSNSPlatformApplication_GCM_basic(t *testing.T) {
	ctx := acctest.Context(t)
	apiKey := acctest.SkipIfEnvVarNotSet(t, "GCM_API_KEY")
	resourceName := "aws_sns_platform_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlatformApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlatformApplicationConfig_gcmBasic(rName, apiKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPlatformApplicationExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "apple_platform_bundle_id"),
					resource.TestCheckNoResourceAttr(resourceName, "apple_platform_team_id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sns", fmt.Sprintf("app/GCM/%s", rName)),
					resource.TestCheckNoResourceAttr(resourceName, "event_delivery_failure_topic_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "event_endpoint_created_topic_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "event_endpoint_deleted_topic_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "event_endpoint_updated_topic_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "failure_feedback_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "platform", "GCM"),
					resource.TestCheckNoResourceAttr(resourceName, "success_feedback_role_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "success_feedback_sample_rate"),
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
}

func TestAccSNSPlatformApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	apiKey := acctest.SkipIfEnvVarNotSet(t, "GCM_API_KEY")
	resourceName := "aws_sns_platform_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlatformApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlatformApplicationConfig_gcmBasic(rName, apiKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlatformApplicationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsns.ResourcePlatformApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSNSPlatformApplication_GCM_allAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	apiKey := acctest.SkipIfEnvVarNotSet(t, "GCM_API_KEY")
	resourceName := "aws_sns_platform_application.test"
	topic0ResourceName := "aws_sns_topic.test.0"
	topic1ResourceName := "aws_sns_topic.test.1"
	role0ResourceName := "aws_iam_role.test.0"
	role1ResourceName := "aws_iam_role.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlatformApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlatformApplicationConfig_gcmAllAttributes(rName, apiKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPlatformApplicationExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "apple_platform_bundle_id"),
					resource.TestCheckNoResourceAttr(resourceName, "apple_platform_team_id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sns", fmt.Sprintf("app/GCM/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "event_delivery_failure_topic_arn", topic0ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "event_endpoint_created_topic_arn", topic1ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "event_endpoint_deleted_topic_arn", topic0ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "event_endpoint_updated_topic_arn", topic1ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "failure_feedback_role_arn", role0ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "platform", "GCM"),
					resource.TestCheckResourceAttrPair(resourceName, "success_feedback_role_arn", role1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "success_feedback_sample_rate", "25"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"platform_credential", "platform_principal"},
			},
			{
				Config: testAccPlatformApplicationConfig_gcmAllAttributesUpdated(rName, apiKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPlatformApplicationExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "apple_platform_bundle_id"),
					resource.TestCheckNoResourceAttr(resourceName, "apple_platform_team_id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sns", fmt.Sprintf("app/GCM/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "event_delivery_failure_topic_arn", topic1ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "event_endpoint_created_topic_arn", topic0ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "event_endpoint_deleted_topic_arn", topic1ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "event_endpoint_updated_topic_arn", topic0ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "failure_feedback_role_arn", role1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "platform", "GCM"),
					resource.TestCheckResourceAttrPair(resourceName, "success_feedback_role_arn", role0ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "success_feedback_sample_rate", "50"),
				),
			},
		},
	})
}

func TestAccSNSPlatformApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	platforms := testAccPlatformApplicationPlatformFromEnv(t, "certificate")
	resourceName := "aws_sns_platform_application.test"

	for _, platform := range platforms { //nolint:paralleltest
		name := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())
		platformPrincipalCheck := resource.TestCheckNoResourceAttr(resourceName, "platform_principal")
		if platform.Principal != "" {
			platformPrincipalCheck = resource.TestCheckResourceAttrSet(resourceName, "platform_principal")
		}

		t.Run(platform.Name, func(*testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.SNSEndpointID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckPlatformApplicationDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccPlatformApplicationConfig_basic(name, platform),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckPlatformApplicationExists(ctx, resourceName),
							acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sns", regexache.MustCompile(fmt.Sprintf("app/%s/%s$", platform.Name, name))),
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
	ctx := acctest.Context(t)
	platforms := testAccPlatformApplicationPlatformFromEnv(t, "certificate")
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

	for _, platform := range platforms { //nolint:paralleltest
		t.Run(platform.Name, func(*testing.T) {
			t.Parallel()

			for _, tc := range testCases {
				t.Run(fmt.Sprintf("%s/%s", platform.Name, tc.AttributeKey), func(*testing.T) {
					name := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())

					resource.Test(t, resource.TestCase{
						PreCheck:                 func() { acctest.PreCheck(ctx, t) },
						ErrorCheck:               acctest.ErrorCheck(t, names.SNSEndpointID),
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						CheckDestroy:             testAccCheckPlatformApplicationDestroy(ctx),
						Steps: []resource.TestStep{
							{
								Config: testAccPlatformApplicationConfig_basicAttribute(name, platform, tc.AttributeKey, tc.AttributeValue),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckPlatformApplicationExists(ctx, resourceName),
									resource.TestCheckResourceAttr(resourceName, tc.AttributeKey, tc.AttributeValue),
								),
							},
							{
								Config: testAccPlatformApplicationConfig_basicAttribute(name, platform, tc.AttributeKey, tc.AttributeValueUpdate),
								Check: resource.ComposeTestCheckFunc(
									testAccCheckPlatformApplicationExists(ctx, resourceName),
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

func TestAccSNSPlatformApplication_basicApnsWithTokenCredentials(t *testing.T) {
	ctx := acctest.Context(t)
	platforms := testAccPlatformApplicationPlatformFromEnv(t, "token")
	resourceName := "aws_sns_platform_application.test"
	applePlatformTeamId := "1111111111"
	updatedApplePlatformTeamId := "2222222222"
	applePlatformBundleId := "com.bundle.name"
	updatedApplePlatformBundleId := "com.bundle2.name2"

	for _, platform := range platforms { //nolint:paralleltest
		if platform.Name == "GCM" {
			continue
		}

		name := fmt.Sprintf("tf-acc-%d", sdkacctest.RandInt())

		t.Run(platform.Name, func(*testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.SNSEndpointID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckPlatformApplicationDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccPlatformApplicationConfig_basicApnsWithTokenCredentials(name, platform, applePlatformTeamId, applePlatformBundleId),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckPlatformApplicationExists(ctx, resourceName),
							resource.TestCheckResourceAttr(resourceName, "name", name),
							resource.TestCheckResourceAttr(resourceName, "platform", platform.Name),
							resource.TestCheckResourceAttrSet(resourceName, "platform_credential"),
							resource.TestCheckResourceAttrSet(resourceName, "platform_principal"),
							resource.TestCheckResourceAttr(resourceName, "apple_platform_team_id", applePlatformTeamId),
							resource.TestCheckResourceAttr(resourceName, "apple_platform_bundle_id", applePlatformBundleId),
						),
					},
					{
						Config: testAccPlatformApplicationConfig_basicApnsWithTokenCredentials(name, platform, updatedApplePlatformTeamId, updatedApplePlatformBundleId),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckPlatformApplicationExists(ctx, resourceName),
							resource.TestCheckResourceAttr(resourceName, "apple_platform_team_id", updatedApplePlatformTeamId),
							resource.TestCheckResourceAttr(resourceName, "apple_platform_bundle_id", updatedApplePlatformBundleId),
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

func testAccCheckPlatformApplicationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSClient(ctx)

		_, err := tfsns.FindPlatformApplicationAttributesByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckPlatformApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sns_platform_application" {
				continue
			}

			_, err := tfsns.FindPlatformApplicationAttributesByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SNS Platform Application %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPlatformApplicationConfig_gcmBasic(rName, credentials string) string {
	return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = %[1]q
  platform            = "GCM"
  platform_credential = %[2]q
}
`, rName, credentials)
}

func testAccPlatformApplicationConfig_gcmAllAttributesBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2

  name = "%[1]s-${count.index}"
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  count = 2

  name = "%[1]s-${count.index}"

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
}

resource "aws_iam_role_policy_attachment" "test" {
  count = 2

  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/CloudWatchLogsFullAccess"
  role       = aws_iam_role.test[count.index].id
}
`, rName)
}

func testAccPlatformApplicationConfig_gcmAllAttributes(rName, credentials string) string {
	return acctest.ConfigCompose(testAccPlatformApplicationConfig_gcmAllAttributesBase(rName), fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = %[1]q
  platform            = "GCM"
  platform_credential = %[2]q

  event_delivery_failure_topic_arn = aws_sns_topic.test[0].arn
  event_endpoint_created_topic_arn = aws_sns_topic.test[1].arn
  event_endpoint_deleted_topic_arn = aws_sns_topic.test[0].arn
  event_endpoint_updated_topic_arn = aws_sns_topic.test[1].arn

  failure_feedback_role_arn = aws_iam_role.test[0].arn
  success_feedback_role_arn = aws_iam_role.test[1].arn

  success_feedback_sample_rate = 25
}
`, rName, credentials))
}

func testAccPlatformApplicationConfig_gcmAllAttributesUpdated(rName, credentials string) string {
	return acctest.ConfigCompose(testAccPlatformApplicationConfig_gcmAllAttributesBase(rName), fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = %[1]q
  platform            = "GCM"
  platform_credential = %[2]q

  event_delivery_failure_topic_arn = aws_sns_topic.test[1].arn
  event_endpoint_created_topic_arn = aws_sns_topic.test[0].arn
  event_endpoint_deleted_topic_arn = aws_sns_topic.test[1].arn
  event_endpoint_updated_topic_arn = aws_sns_topic.test[0].arn

  failure_feedback_role_arn = aws_iam_role.test[1].arn
  success_feedback_role_arn = aws_iam_role.test[0].arn

  success_feedback_sample_rate = 50
}
`, rName, credentials))
}

func testAccPlatformApplicationConfig_basic(name string, platform *testAccPlatformApplicationPlatform) string {
	if platform.Principal == "" {
		return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = %[1]q
  platform            = %[2]q
  platform_credential = %[3]s
}
`, name, platform.Name, platform.Credential)
	}
	return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = %[1]q
  platform            = %[2]q
  platform_credential = %[3]s
  platform_principal  = %[4]s
}
`, name, platform.Name, platform.Credential, platform.Principal)
}

func testAccPlatformApplicationConfig_basicAttribute(name string, platform *testAccPlatformApplicationPlatform, attributeKey, attributeValue string) string {
	if platform.Principal == "" {
		return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = %[1]q
  platform            = %[2]q
  platform_credential = %[3]s
  %[4]s               = %[5]q
}
`, name, platform.Name, platform.Credential, attributeKey, attributeValue)
	}
	return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                = %[1]q
  platform            = %[2]q
  platform_credential = %[3]s
  platform_principal  = %[4]s
  %[5]s               = %[6]q
}
`, name, platform.Name, platform.Credential, platform.Principal, attributeKey, attributeValue)
}

func testAccPlatformApplicationConfig_basicApnsWithTokenCredentials(name string, platform *testAccPlatformApplicationPlatform, applePlatformTeamId string, applePlatformBundleId string) string {
	return fmt.Sprintf(`
resource "aws_sns_platform_application" "test" {
  name                     = %[1]q
  platform                 = %[2]q
  platform_credential      = %[3]s
  platform_principal       = %[4]s
  apple_platform_team_id   = %[5]q
  apple_platform_bundle_id = %[6]q
}
`, name, platform.Name, platform.Credential, platform.Principal, applePlatformTeamId, applePlatformBundleId)
}
