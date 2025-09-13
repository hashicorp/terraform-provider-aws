// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfpinpoint "github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
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

type testAccAPNSVoipSandboxChannelCertConfiguration struct {
	Certificate string
	PrivateKey  string
}

type testAccAPNSVoipSandboxChannelTokenConfiguration struct {
	BundleId   string
	TeamId     string
	TokenKey   string
	TokenKeyId string
}

func testAccAPNSVoIPSandboxChannelCertConfigurationFromEnv(t *testing.T) *testAccAPNSVoipSandboxChannelCertConfiguration {
	var conf *testAccAPNSVoipSandboxChannelCertConfiguration
	if os.Getenv("APNS_VOIP_CERTIFICATE") != "" {
		if os.Getenv("APNS_VOIP_CERTIFICATE_PRIVATE_KEY") == "" {
			t.Fatalf("APNS_VOIP_CERTIFICATE set but missing APNS_VOIP_CERTIFICATE_PRIVATE_KEY")
		}

		conf = &testAccAPNSVoipSandboxChannelCertConfiguration{
			Certificate: fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_VOIP_CERTIFICATE"))),
			PrivateKey:  fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_VOIP_CERTIFICATE_PRIVATE_KEY"))),
		}
	}

	if conf == nil {
		t.Skipf("Pinpoint certificate credentials envs are missing, skipping test")
	}

	return conf
}

func testAccAPNSVoIPSandboxChannelTokenConfigurationFromEnv(t *testing.T) *testAccAPNSVoipSandboxChannelTokenConfiguration {
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

	conf := testAccAPNSVoipSandboxChannelTokenConfiguration{
		BundleId:   strconv.Quote(strings.TrimSpace(os.Getenv("APNS_VOIP_BUNDLE_ID"))),
		TeamId:     strconv.Quote(strings.TrimSpace(os.Getenv("APNS_VOIP_TEAM_ID"))),
		TokenKey:   fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_VOIP_TOKEN_KEY"))),
		TokenKeyId: strconv.Quote(strings.TrimSpace(os.Getenv("APNS_VOIP_TOKEN_KEY_ID"))),
	}

	return &conf
}

func TestAccPinpointAPNSVoIPSandboxChannel_basicCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.APNSVoipSandboxChannelResponse
	resourceName := "aws_pinpoint_apns_voip_sandbox_channel.test_channel"

	configuration := testAccAPNSVoIPSandboxChannelCertConfigurationFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPNSVoIPSandboxChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPNSVoIPSandboxChannelConfig_basicCertificate(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSVoIPSandboxChannelExists(ctx, resourceName, &channel),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrCertificate, names.AttrPrivateKey},
			},
			{
				Config: testAccAPNSVoIPSandboxChannelConfig_basicCertificate(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSVoIPSandboxChannelExists(ctx, resourceName, &channel),
				),
			},
		},
	})
}

func TestAccPinpointAPNSVoIPSandboxChannel_basicToken(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.APNSVoipSandboxChannelResponse
	resourceName := "aws_pinpoint_apns_voip_sandbox_channel.test_channel"

	configuration := testAccAPNSVoIPSandboxChannelTokenConfigurationFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPNSVoIPSandboxChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPNSVoIPSandboxChannelConfig_basicToken(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSVoIPSandboxChannelExists(ctx, resourceName, &channel),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"team_id", "bundle_id", "token_key", "token_key_id"},
			},
			{
				Config: testAccAPNSVoIPSandboxChannelConfig_basicToken(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSVoIPSandboxChannelExists(ctx, resourceName, &channel),
				),
			},
		},
	})
}

func testAccCheckAPNSVoIPSandboxChannelExists(ctx context.Context, n string, channel *awstypes.APNSVoipSandboxChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint APNS VoIP Sandbox Channel with that Application ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointClient(ctx)

		output, err := tfpinpoint.FindAPNSVoIPSandboxChannelByApplicationId(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*channel = *output

		return nil
	}
}

func testAccCheckAPNSVoIPSandboxChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_apns_voip_sandbox_channel" {
				continue
			}

			_, err := tfpinpoint.FindAPNSVoIPSandboxChannelByApplicationId(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Pinpoint APNS VoIP Sandbox Channel %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAPNSVoIPSandboxChannelConfig_basicCertificate(conf *testAccAPNSVoipSandboxChannelCertConfiguration) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_apns_voip_sandbox_channel" "test_channel" {
  application_id                = aws_pinpoint_app.test_app.application_id
  enabled                       = false
  default_authentication_method = "CERTIFICATE"
  certificate                   = %s
  private_key                   = %s
}
`, conf.Certificate, conf.PrivateKey)
}

func testAccAPNSVoIPSandboxChannelConfig_basicToken(conf *testAccAPNSVoipSandboxChannelTokenConfiguration) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_apns_voip_sandbox_channel" "test_channel" {
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
