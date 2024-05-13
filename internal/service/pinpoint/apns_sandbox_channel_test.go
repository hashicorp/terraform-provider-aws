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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

/**
 Before running this test, one of the following two ENV variables set must be defined. See here for details:
 https://docs.aws.amazon.com/pinpoint/latest/userguide/channels-mobile-manage.html

 * Key Configuration (ref. https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/establishing_a_token-based_connection_to_apns )
 APNS_SANDBOX_BUNDLE_ID    - APNs Bundle ID
 APNS_SANDBOX_TEAM_ID      - APNs Team ID
 APNS_SANDBOX_TOKEN_KEY    - Token key file content (.p8 file)
 APNS_SANDBOX_TOKEN_KEY_ID - APNs Token Key ID

 * Certificate Configuration (ref. https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/establishing_a_certificate-based_connection_to_apns )
 APNS_SANDBOX_CERTIFICATE             - APNs Certificate content (.pem file content)
 APNS_SANDBOX_CERTIFICATE_PRIVATE_KEY - APNs Certificate Private Key File content
**/

type testAccAPNSSandboxChannelCertConfiguration struct {
	Certificate string
	PrivateKey  string
}

type testAccAPNSSandboxChannelTokenConfiguration struct {
	BundleId   string
	TeamId     string
	TokenKey   string
	TokenKeyId string
}

func testAccAPNSSandboxChannelCertConfigurationFromEnv(t *testing.T) *testAccAPNSSandboxChannelCertConfiguration {
	var conf *testAccAPNSSandboxChannelCertConfiguration
	if os.Getenv("APNS_SANDBOX_CERTIFICATE") != "" {
		if os.Getenv("APNS_SANDBOX_CERTIFICATE_PRIVATE_KEY") == "" {
			t.Fatalf("APNS_SANDBOX_CERTIFICATE set but missing APNS_SANDBOX_CERTIFICATE_PRIVATE_KEY")
		}

		conf = &testAccAPNSSandboxChannelCertConfiguration{
			Certificate: fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_SANDBOX_CERTIFICATE"))),
			PrivateKey:  fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_SANDBOX_CERTIFICATE_PRIVATE_KEY"))),
		}
	}

	if conf == nil {
		t.Skipf("Pinpoint certificate credentials envs are missing, skipping test")
	}

	return conf
}

func testAccAPNSSandboxChannelTokenConfigurationFromEnv(t *testing.T) *testAccAPNSSandboxChannelTokenConfiguration {
	if os.Getenv("APNS_SANDBOX_BUNDLE_ID") == "" {
		t.Skipf("APNS_SANDBOX_BUNDLE_ID env is missing, skipping test")
	}

	if os.Getenv("APNS_SANDBOX_TEAM_ID") == "" {
		t.Skipf("APNS_SANDBOX_TEAM_ID env is missing, skipping test")
	}

	if os.Getenv("APNS_SANDBOX_TOKEN_KEY") == "" {
		t.Skipf("APNS_SANDBOX_TOKEN_KEY env is missing, skipping test")
	}

	if os.Getenv("APNS_SANDBOX_TOKEN_KEY_ID") == "" {
		t.Skipf("APNS_SANDBOX_TOKEN_KEY_ID env is missing, skipping test")
	}

	conf := testAccAPNSSandboxChannelTokenConfiguration{
		BundleId:   strconv.Quote(strings.TrimSpace(os.Getenv("APNS_SANDBOX_BUNDLE_ID"))),
		TeamId:     strconv.Quote(strings.TrimSpace(os.Getenv("APNS_SANDBOX_TEAM_ID"))),
		TokenKey:   fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_SANDBOX_TOKEN_KEY"))),
		TokenKeyId: strconv.Quote(strings.TrimSpace(os.Getenv("APNS_SANDBOX_TOKEN_KEY_ID"))),
	}

	return &conf
}

func TestAccPinpointAPNSSandboxChannel_basicCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	var channel pinpoint.APNSSandboxChannelResponse
	resourceName := "aws_pinpoint_apns_sandbox_channel.test_channel"

	configuration := testAccAPNSSandboxChannelCertConfigurationFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPNSSandboxChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPNSSandboxChannelConfig_basicCertificate(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSSandboxChannelExists(ctx, resourceName, &channel),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrCertificate, names.AttrPrivateKey},
			},
			{
				Config: testAccAPNSSandboxChannelConfig_basicCertificate(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSSandboxChannelExists(ctx, resourceName, &channel),
				),
			},
		},
	})
}

func TestAccPinpointAPNSSandboxChannel_basicToken(t *testing.T) {
	ctx := acctest.Context(t)
	var channel pinpoint.APNSSandboxChannelResponse
	resourceName := "aws_pinpoint_apns_sandbox_channel.test_channel"

	configuration := testAccAPNSSandboxChannelTokenConfigurationFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPNSSandboxChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPNSSandboxChannelConfig_basicToken(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSSandboxChannelExists(ctx, resourceName, &channel),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"team_id", "bundle_id", "token_key", "token_key_id"},
			},
			{
				Config: testAccAPNSSandboxChannelConfig_basicToken(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPNSSandboxChannelExists(ctx, resourceName, &channel),
				),
			},
		},
	})
}

func testAccCheckAPNSSandboxChannelExists(ctx context.Context, n string, channel *pinpoint.APNSSandboxChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint APNs Channel with that Application ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		// Check if the app exists
		params := &pinpoint.GetApnsSandboxChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetApnsSandboxChannelWithContext(ctx, params)

		if err != nil {
			return err
		}

		*channel = *output.APNSSandboxChannelResponse

		return nil
	}
}

func testAccAPNSSandboxChannelConfig_basicCertificate(conf *testAccAPNSSandboxChannelCertConfiguration) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_apns_sandbox_channel" "test_channel" {
  application_id                = aws_pinpoint_app.test_app.application_id
  enabled                       = false
  default_authentication_method = "CERTIFICATE"
  certificate                   = %s
  private_key                   = %s
}
`, conf.Certificate, conf.PrivateKey)
}

func testAccAPNSSandboxChannelConfig_basicToken(conf *testAccAPNSSandboxChannelTokenConfiguration) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_apns_sandbox_channel" "test_channel" {
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

func testAccCheckAPNSSandboxChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_apns_sandbox_channel" {
				continue
			}

			// Check if the channel exists
			params := &pinpoint.GetApnsSandboxChannelInput{
				ApplicationId: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetApnsSandboxChannelWithContext(ctx, params)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
					continue
				}
				return err
			}
			return fmt.Errorf("APNs Sandbox Channel exists when it should be destroyed!")
		}

		return nil
	}
}
