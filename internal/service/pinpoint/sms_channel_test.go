// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpoint "github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.SMSChannelResponse
	resourceName := "aws_pinpoint_sms_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMSChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSMSChannelConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMSChannelExists(ctx, t, resourceName, &channel),
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
					testAccCheckSMSChannelExists(ctx, t, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccPinpointSMSChannel_full(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.SMSChannelResponse
	resourceName := "aws_pinpoint_sms_channel.test"
	senderId := "1234"
	shortCode := "5678"
	newShortCode := "7890"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMSChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSMSChannelConfig_full(senderId, shortCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMSChannelExists(ctx, t, resourceName, &channel),
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
					testAccCheckSMSChannelExists(ctx, t, resourceName, &channel),
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
	var channel awstypes.SMSChannelResponse
	resourceName := "aws_pinpoint_sms_channel.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMSChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSMSChannelConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMSChannelExists(ctx, t, resourceName, &channel),
					acctest.CheckSDKResourceDisappears(ctx, t, tfpinpoint.ResourceSMSChannel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSMSChannelExists(ctx context.Context, t *testing.T, n string, channel *awstypes.SMSChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint SMS Channel with that application ID exists")
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointClient(ctx)

		output, err := tfpinpoint.FindSMSChannelByApplicationId(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*channel = *output

		return nil
	}
}

func testAccCheckSMSChannelDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_sms_channel" {
				continue
			}

			_, err := tfpinpoint.FindSMSChannelByApplicationId(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Pinpoint SMS Channel %s still exists", rs.Primary.ID)
		}

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
