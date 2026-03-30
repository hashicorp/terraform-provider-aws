// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpoint "github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
	"github.com/hashicorp/terraform-provider-aws/names"
)

/**
 Before running this test, the following two ENV variables must be set:

 ADM_CLIENT_ID     - Amazon ADM OAuth Credentials Client ID
 ADM_CLIENT_SECRET - Amazon ADM OAuth Credentials Client Secret
**/

type testAccADMChannelConfiguration struct {
	ClientID     string
	ClientSecret string
}

func testAccADMChannelConfigurationFromEnv(t *testing.T) *testAccADMChannelConfiguration {
	if os.Getenv("ADM_CLIENT_ID") == "" {
		t.Skipf("ADM_CLIENT_ID ENV is missing")
	}

	if os.Getenv("ADM_CLIENT_SECRET") == "" {
		t.Skipf("ADM_CLIENT_SECRET ENV is missing")
	}

	conf := testAccADMChannelConfiguration{
		ClientID:     os.Getenv("ADM_CLIENT_ID"),
		ClientSecret: os.Getenv("ADM_CLIENT_SECRET"),
	}

	return &conf
}

func TestAccPinpointADMChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.ADMChannelResponse
	resourceName := "aws_pinpoint_adm_channel.channel"

	config := testAccADMChannelConfigurationFromEnv(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckADMChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccADMChannelConfig_basic(config),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADMChannelExists(ctx, t, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrClientID, names.AttrClientSecret},
			},
			{
				Config: testAccADMChannelConfig_basic(config),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADMChannelExists(ctx, t, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckADMChannelExists(ctx context.Context, t *testing.T, n string, channel *awstypes.ADMChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint ADM channel with that Application ID exists")
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointClient(ctx)

		output, err := tfpinpoint.FindADMChannelByApplicationId(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*channel = *output

		return nil
	}
}

func testAccCheckADMChannelDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_adm_channel" {
				continue
			}

			_, err := tfpinpoint.FindADMChannelByApplicationId(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Pinpoint ADM Channel %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccADMChannelConfig_basic(conf *testAccADMChannelConfiguration) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_adm_channel" "channel" {
  application_id = aws_pinpoint_app.test_app.application_id

  client_id     = "%s"
  client_secret = "%s"
  enabled       = false
}
`, conf.ClientID, conf.ClientSecret)
}
