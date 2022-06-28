package pinpoint_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

/**
 Before running this test, one of the following two ENV variables set must be defined. See here for details:
 https://docs.aws.amazon.com/pinpoint/latest/userguide/channels-mobile-manage.html

 * Key Configuration (ref. https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/establishing_a_token-based_connection_to_apns )
 APNS_BUNDLE_ID    - APNs Bundle ID
 APNS_TEAM_ID      - APNs Team ID
 APNS_TOKEN_KEY    - Token key file content (.p8 file)
 APNS_TOKEN_KEY_ID - APNs Token Key ID

 * Certificate Configuration (ref. https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/establishing_a_certificate-based_connection_to_apns )
 APNS_CERTIFICATE             - APNs Certificate content (.pem file content)
 APNS_CERTIFICATE_PRIVATE_KEY - APNs Certificate Private Key File content
**/

type testAccAPNSChannelCertConfiguration struct {
	Certificate string
	PrivateKey  string
}

type testAccAPNSChannelTokenConfiguration struct {
	BundleId   string
	TeamId     string
	TokenKey   string
	TokenKeyId string
}

func testAccAPNSChannelCertConfigurationFromEnv(t *testing.T) *testAccAPNSChannelCertConfiguration {
	var conf *testAccAPNSChannelCertConfiguration
	if os.Getenv("APNS_CERTIFICATE") != "" {
		if os.Getenv("APNS_CERTIFICATE_PRIVATE_KEY") == "" {
			t.Fatalf("APNS_CERTIFICATE set but missing APNS_CERTIFICATE_PRIVATE_KEY")
		}

		conf = &testAccAPNSChannelCertConfiguration{
			Certificate: fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_CERTIFICATE"))),
			PrivateKey:  fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_CERTIFICATE_PRIVATE_KEY"))),
		}
	}

	if conf == nil {
		t.Skipf("Pinpoint certificate credentials envs are missing, skipping test")
	}

	return conf
}

func testAccAPNSChannelTokenConfigurationFromEnv(t *testing.T) *testAccAPNSChannelTokenConfiguration {
	if os.Getenv("APNS_BUNDLE_ID") == "" {
		t.Skipf("APNS_BUNDLE_ID env is missing, skipping test")
	}

	if os.Getenv("APNS_TEAM_ID") == "" {
		t.Skipf("APNS_TEAM_ID env is missing, skipping test")
	}

	if os.Getenv("APNS_TOKEN_KEY") == "" {
		t.Skipf("APNS_TOKEN_KEY env is missing, skipping test")
	}

	if os.Getenv("APNS_TOKEN_KEY_ID") == "" {
		t.Skipf("APNS_TOKEN_KEY_ID env is missing, skipping test")
	}

	conf := testAccAPNSChannelTokenConfiguration{
		BundleId:   strconv.Quote(strings.TrimSpace(os.Getenv("APNS_BUNDLE_ID"))),
		TeamId:     strconv.Quote(strings.TrimSpace(os.Getenv("APNS_TEAM_ID"))),
		TokenKey:   fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_TOKEN_KEY"))),
		TokenKeyId: strconv.Quote(strings.TrimSpace(os.Getenv("APNS_TOKEN_KEY_ID"))),
	}

	return &conf
}

func TestAccPinpointAPNSChannel_basicCertificate(t *testing.T) {
	var channel pinpoint.APNSChannelResponse
	resourceName := "aws_pinpoint_apns_channel.test_apns_channel"

	configuration := testAccAPNSChannelCertConfigurationFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPNSChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPNSChannelConfig_basicCertificate(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSChannelExists(resourceName, &channel),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate", "private_key"},
			},
			{
				Config: testAccAPNSChannelConfig_basicCertificate(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSChannelExists(resourceName, &channel),
				),
			},
		},
	})
}

func TestAccPinpointAPNSChannel_basicToken(t *testing.T) {
	var channel pinpoint.APNSChannelResponse
	resourceName := "aws_pinpoint_apns_channel.test_apns_channel"

	configuration := testAccAPNSChannelTokenConfigurationFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPNSChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPNSChannelConfig_basicToken(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSChannelExists(resourceName, &channel),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"team_id", "bundle_id", "token_key", "token_key_id"},
			},
			{
				Config: testAccAPNSChannelConfig_basicToken(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSChannelExists(resourceName, &channel),
				),
			},
		},
	})
}

func testAccCheckAPNSChannelExists(n string, channel *pinpoint.APNSChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint APNs Channel with that Application ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

		// Check if the app exists
		params := &pinpoint.GetApnsChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetApnsChannel(params)

		if err != nil {
			return err
		}

		*channel = *output.APNSChannelResponse

		return nil
	}
}

func testAccAPNSChannelConfig_basicCertificate(conf *testAccAPNSChannelCertConfiguration) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_apns_channel" "test_apns_channel" {
  application_id                = aws_pinpoint_app.test_app.application_id
  enabled                       = false
  default_authentication_method = "CERTIFICATE"
  certificate                   = %s
  private_key                   = %s
}
`, conf.Certificate, conf.PrivateKey)
}

func testAccAPNSChannelConfig_basicToken(conf *testAccAPNSChannelTokenConfiguration) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_apns_channel" "test_apns_channel" {
  application_id = aws_pinpoint_app.test_app.application_id
  enabled        = false

  default_authentication_method = "TOKEN"

  bundle_id    = %s
  team_id      = %s
  token_key    = %s
  token_key_id = %s
}
`, conf.BundleId, conf.TeamId, conf.TokenKey, conf.TokenKeyId)
}

func testAccCheckAPNSChannelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_apns_channel" {
			continue
		}

		// Check if the channel exists
		params := &pinpoint.GetApnsChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetApnsChannel(params)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
				continue
			}
			return err
		}
		return fmt.Errorf("APNs Channel exists when it should be destroyed!")
	}

	return nil
}
