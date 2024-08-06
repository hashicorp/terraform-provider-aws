// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfpinpoint "github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var channel pinpoint.SMSChannelResponse
	resourceName := "aws_pinpoint_sms_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMSChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMSChannelConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMSChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// There can be a delay before these Computed values are returned
				// e.g. 0 on Create -> Read, 20 on Import
				// These seem non-critical for other Terraform resource references,
				// so ignoring them for now, but we can likely adjust the Read function
				// to wait until they are available on creation with retry logic.
				ImportStateVerifyIgnore: []string{
					"promotional_messages_per_second",
					"transactional_messages_per_second",
				},
			},
			{
				Config: testAccSMSChannelConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMSChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccPinpointSMSChannel_full(t *testing.T) {
	ctx := acctest.Context(t)
	var channel pinpoint.SMSChannelResponse
	resourceName := "aws_pinpoint_sms_channel.test"
	senderId := "1234"
	shortCode := "5678"
	newShortCode := "7890"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMSChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMSChannelConfig_full(senderId, shortCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMSChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "sender_id", senderId),
					resource.TestCheckResourceAttr(resourceName, "short_code", shortCode),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "promotional_messages_per_second"),
					resource.TestCheckResourceAttrSet(resourceName, "transactional_messages_per_second"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// There can be a delay before these Computed values are returned
				// e.g. 0 on Create -> Read, 20 on Import
				// These seem non-critical for other Terraform resource references,
				// so ignoring them for now, but we can likely adjust the Read function
				// to wait until they are available on creation with retry logic.
				ImportStateVerifyIgnore: []string{
					"promotional_messages_per_second",
					"transactional_messages_per_second",
				},
			},
			{
				Config: testAccSMSChannelConfig_full(senderId, newShortCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMSChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "sender_id", senderId),
					resource.TestCheckResourceAttr(resourceName, "short_code", newShortCode),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "promotional_messages_per_second"),
					resource.TestCheckResourceAttrSet(resourceName, "transactional_messages_per_second"),
				),
			},
		},
	})
}

func TestAccPinpointSMSChannel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var channel pinpoint.SMSChannelResponse
	resourceName := "aws_pinpoint_sms_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMSChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMSChannelConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMSChannelExists(ctx, resourceName, &channel),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfpinpoint.ResourceSMSChannel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSMSChannelExists(ctx context.Context, n string, channel *pinpoint.SMSChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint SMS Channel with that application ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		// Check if the app exists
		params := &pinpoint.GetSmsChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetSmsChannelWithContext(ctx, params)

		if err != nil {
			return err
		}

		*channel = *output.SMSChannelResponse

		return nil
	}
}

const testAccSMSChannelConfig_basic = `
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_sms_channel" "test" {
  application_id = aws_pinpoint_app.test_app.application_id
}
`

func testAccSMSChannelConfig_full(senderId, shortCode string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_sms_channel" "test" {
  application_id = aws_pinpoint_app.test_app.application_id
  enabled        = "false"
  sender_id      = "%s"
  short_code     = "%s"
}
`, senderId, shortCode)
}

func testAccCheckSMSChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_sms_channel" {
				continue
			}

			// Check if the event stream exists
			params := &pinpoint.GetSmsChannelInput{
				ApplicationId: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetSmsChannelWithContext(ctx, params)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
					continue
				}
				return err
			}
			return fmt.Errorf("SMS Channel exists when it should be destroyed!")
		}

		return nil
	}
}
