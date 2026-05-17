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

func TestAccPinpointBaiduChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.BaiduChannelResponse
	resourceName := "aws_pinpoint_baidu_channel.channel"

	apiKey := "123"
	apikeyUpdated := "234"
	secretKey := "456"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBaiduChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBaiduChannelConfig_basic(apiKey, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaiduChannelExists(ctx, t, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "api_key", apiKey),
					resource.TestCheckResourceAttr(resourceName, names.AttrSecretKey, secretKey),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key", names.AttrSecretKey},
			},
			{
				Config: testAccBaiduChannelConfig_basic(apikeyUpdated, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaiduChannelExists(ctx, t, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "api_key", apikeyUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrSecretKey, secretKey),
				),
			},
		},
	})
}

func testAccCheckBaiduChannelExists(ctx context.Context, t *testing.T, n string, channel *awstypes.BaiduChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint Baidu channel with that Application ID exists")
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointClient(ctx)

		output, err := tfpinpoint.FindBaiduChannelByApplicationId(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*channel = *output

		return nil
	}
}

func testAccCheckBaiduChannelDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_baidu_channel" {
				continue
			}

			_, err := tfpinpoint.FindBaiduChannelByApplicationId(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Pinpoint Baidu Channel %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBaiduChannelConfig_basic(apiKey, secretKey string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_baidu_channel" "channel" {
  application_id = aws_pinpoint_app.test_app.application_id

  enabled    = "false"
  api_key    = "%s"
  secret_key = "%s"
}
`, apiKey, secretKey)
}
