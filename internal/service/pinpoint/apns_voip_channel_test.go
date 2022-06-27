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
 APNS_VOIP_BUNDLE_ID    - APNs Bundle ID
 APNS_VOIP_TEAM_ID      - APNs Team ID
 APNS_VOIP_TOKEN_KEY    - Token key file content (.p8 file)
 APNS_VOIP_TOKEN_KEY_ID - APNs Token Key ID

 * Certificate Configuration (ref. https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/establishing_a_certificate-based_connection_to_apns - Select "VoIP Services Certificate" )
 APNS_VOIP_CERTIFICATE             - APNs Certificate content (.pem file content)
 APNS_VOIP_CERTIFICATE_PRIVATE_KEY - APNs Certificate Private Key File content
**/

type testAccAPNSVoipChannelCertConfiguration struct {
	Certificate string
	PrivateKey  string
}

type testAccAPNSVoipChannelTokenConfiguration struct {
	BundleId   string
	TeamId     string
	TokenKey   string
	TokenKeyId string
}

func testAccAPNSVoIPChannelCertConfigurationFromEnv(t *testing.T) *testAccAPNSVoipChannelCertConfiguration {
	var conf *testAccAPNSVoipChannelCertConfiguration
	if os.Getenv("APNS_VOIP_CERTIFICATE") != "" {
		if os.Getenv("APNS_VOIP_CERTIFICATE_PRIVATE_KEY") == "" {
			t.Fatalf("APNS_VOIP_CERTIFICATE set but missing APNS_VOIP_CERTIFICATE_PRIVATE_KEY")
		}

		conf = &testAccAPNSVoipChannelCertConfiguration{
			Certificate: fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_VOIP_CERTIFICATE"))),
			PrivateKey:  fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_VOIP_CERTIFICATE_PRIVATE_KEY"))),
		}
	}

	if conf == nil {
		t.Skipf("Pinpoint certificate credentials envs are missing, skipping test")
	}

	return conf
}

func testAccAPNSVoIPChannelTokenConfigurationFromEnv(t *testing.T) *testAccAPNSVoipChannelTokenConfiguration {
	if os.Getenv("APNS_VOIP_BUNDLE_ID") == "" {
		t.Skipf("APNS_VOIP_BUNDLE_ID env is missing, skipping test")
	}

	if os.Getenv("APNS_VOIP_TEAM_ID") == "" {
		t.Skipf("APNS_VOIP_TEAM_ID env is missing, skipping test")
	}

	if os.Getenv("APNS_VOIP_TOKEN_KEY") == "" {
		t.Skipf("APNS_VOIP_TOKEN_KEY env is missing, skipping test")
	}

	if os.Getenv("APNS_VOIP_TOKEN_KEY_ID") == "" {
		t.Skipf("APNS_VOIP_TOKEN_KEY_ID env is missing, skipping test")
	}

	conf := testAccAPNSVoipChannelTokenConfiguration{
		BundleId:   strconv.Quote(strings.TrimSpace(os.Getenv("APNS_VOIP_BUNDLE_ID"))),
		TeamId:     strconv.Quote(strings.TrimSpace(os.Getenv("APNS_VOIP_TEAM_ID"))),
		TokenKey:   fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_VOIP_TOKEN_KEY"))),
		TokenKeyId: strconv.Quote(strings.TrimSpace(os.Getenv("APNS_VOIP_TOKEN_KEY_ID"))),
	}

	return &conf
}

func TestAccPinpointAPNSVoIPChannel_basicCertificate(t *testing.T) {
	var channel pinpoint.APNSVoipChannelResponse
	resourceName := "aws_pinpoint_apns_voip_channel.test_channel"

	configuration := testAccAPNSVoIPChannelCertConfigurationFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPNSVoIPChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPNSVoIPChannelConfig_basicCertificate(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSVoIPChannelExists(resourceName, &channel),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate", "private_key"},
			},
			{
				Config: testAccAPNSVoIPChannelConfig_basicCertificate(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSVoIPChannelExists(resourceName, &channel),
				),
			},
		},
	})
}

func TestAccPinpointAPNSVoIPChannel_basicToken(t *testing.T) {
	var channel pinpoint.APNSVoipChannelResponse
	resourceName := "aws_pinpoint_apns_voip_channel.test_channel"

	configuration := testAccAPNSVoIPChannelTokenConfigurationFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPNSVoIPChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPNSVoIPChannelConfig_basicToken(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSVoIPChannelExists(resourceName, &channel),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"team_id", "bundle_id", "token_key", "token_key_id"},
			},
			{
				Config: testAccAPNSVoIPChannelConfig_basicToken(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSVoIPChannelExists(resourceName, &channel),
				),
			},
		},
	})
}

func testAccCheckAPNSVoIPChannelExists(n string, channel *pinpoint.APNSVoipChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint APNs Voip Channel with that Application ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

		// Check if the app exists
		params := &pinpoint.GetApnsVoipChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetApnsVoipChannel(params)

		if err != nil {
			return err
		}

		*channel = *output.APNSVoipChannelResponse

		return nil
	}
}

func testAccAPNSVoIPChannelConfig_basicCertificate(conf *testAccAPNSVoipChannelCertConfiguration) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_apns_voip_channel" "test_channel" {
  application_id                = aws_pinpoint_app.test_app.application_id
  enabled                       = false
  default_authentication_method = "CERTIFICATE"
  certificate                   = %s
  private_key                   = %s
}
`, conf.Certificate, conf.PrivateKey)
}

func testAccAPNSVoIPChannelConfig_basicToken(conf *testAccAPNSVoipChannelTokenConfiguration) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_apns_voip_channel" "test_channel" {
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

func testAccCheckAPNSVoIPChannelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_apns_voip_channel" {
			continue
		}

		// Check if the channel exists
		params := &pinpoint.GetApnsVoipChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetApnsVoipChannel(params)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
				continue
			}
			return err
		}
		return fmt.Errorf("APNs Voip Channel exists when it should be destroyed!")
	}

	return nil
}
