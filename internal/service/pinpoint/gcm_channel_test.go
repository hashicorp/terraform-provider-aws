// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"fmt"
	"os"
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
 Before running this test, the following ENV variable must be set:

 GCM_API_KEY - Google Cloud Messaging Api Key
**/

func TestAccPinpointGCMChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var channel pinpoint.GCMChannelResponse
	resourceName := "aws_pinpoint_gcm_channel.test_gcm_channel"

	if os.Getenv("GCM_API_KEY") == "" {
		t.Skipf("GCM_API_KEY env missing, skip test")
	}

	apiKey := os.Getenv("GCM_API_KEY")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGCMChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGCMChannelConfig_basic(apiKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGCMChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
			{
				Config: testAccGCMChannelConfig_basic(apiKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGCMChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckGCMChannelExists(ctx context.Context, n string, channel *pinpoint.GCMChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint GCM Channel with that application ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		// Check if the app exists
		params := &pinpoint.GetGcmChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetGcmChannelWithContext(ctx, params)

		if err != nil {
			return err
		}

		*channel = *output.GCMChannelResponse

		return nil
	}
}

func testAccGCMChannelConfig_basic(apiKey string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_gcm_channel" "test_gcm_channel" {
  application_id = aws_pinpoint_app.test_app.application_id
  enabled        = "false"
  api_key        = "%s"
}
`, apiKey)
}

func testAccCheckGCMChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_gcm_channel" {
				continue
			}

			// Check if the event stream exists
			params := &pinpoint.GetGcmChannelInput{
				ApplicationId: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetGcmChannelWithContext(ctx, params)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
					continue
				}
				return err
			}
			return fmt.Errorf("GCM Channel exists when it should be destroyed!")
		}

		return nil
	}
}
