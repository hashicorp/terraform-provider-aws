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
 Before running this test, the following ENV variable must be set:

 GCM_API_KEY - Google Cloud Messaging Api Key
 GCM_SERVICE_JSON_FILE - Path to a valid Google Cloud Messaging Token File
**/

func TestAccPinpointGCMChannel_basicAPIKey(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.GCMChannelResponse
	resourceName := "aws_pinpoint_gcm_channel.test_gcm_channel"

	if os.Getenv("GCM_API_KEY") == "" {
		t.Skipf("GCM_API_KEY env missing, skip test")
	}

	apiKey := os.Getenv("GCM_API_KEY")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGCMChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGCMChannelConfig_basicAPIKey(apiKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGCMChannelExists(ctx, t, resourceName, &channel),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, "aws_pinpoint_app.test_app", names.AttrApplicationID),
					resource.TestCheckResourceAttr(resourceName, "default_authentication_method", tfpinpoint.DefaultAuthenticationMethodKey),
					resource.TestCheckResourceAttr(resourceName, "api_key", apiKey),
					resource.TestCheckNoResourceAttr(resourceName, "service_json"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"api_key",
				},
			},
		},
	})
}

func TestAccPinpointGCMChannel_apiKeyAuthMethod(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.GCMChannelResponse
	resourceName := "aws_pinpoint_gcm_channel.test_gcm_channel"

	if os.Getenv("GCM_API_KEY") == "" {
		t.Skipf("GCM_API_KEY env missing, skip test")
	}

	apiKey := os.Getenv("GCM_API_KEY")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGCMChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGCMChannelConfig_apiKeyAuthMethod(apiKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGCMChannelExists(ctx, t, resourceName, &channel),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, "aws_pinpoint_app.test_app", names.AttrApplicationID),
					resource.TestCheckResourceAttr(resourceName, "default_authentication_method", tfpinpoint.DefaultAuthenticationMethodKey),
					resource.TestCheckResourceAttr(resourceName, "api_key", apiKey),
					resource.TestCheckNoResourceAttr(resourceName, "service_json"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"api_key",
				},
			},
		},
	})
}

func TestAccPinpointGCMChannel_tokenAuthMethod(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.GCMChannelResponse
	resourceName := "aws_pinpoint_gcm_channel.test_gcm_channel"

	if os.Getenv("GCM_SERVICE_JSON_FILE") == "" {
		t.Skipf("GCM_SERVICE_JSON_FILE env missing, skip test")
	}

	serviceJsonFile := os.Getenv("GCM_SERVICE_JSON_FILE")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGCMChannelDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGCMChannelConfig_tokenAuthMethod(serviceJsonFile),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGCMChannelExists(ctx, t, resourceName, &channel),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, "aws_pinpoint_app.test_app", names.AttrApplicationID),
					resource.TestCheckResourceAttr(resourceName, "default_authentication_method", tfpinpoint.DefaultAuthenticationMethodToken),
					resource.TestCheckNoResourceAttr(resourceName, "api_key"),
					resource.TestCheckResourceAttrSet(resourceName, "service_json"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"service_json",
				},
			},
		},
	})
}

func testAccCheckGCMChannelExists(ctx context.Context, t *testing.T, n string, channel *awstypes.GCMChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint GCM Channel with that application ID exists")
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointClient(ctx)

		output, err := tfpinpoint.FindGCMChannelByApplicationId(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*channel = *output

		return nil
	}
}

func testAccCheckGCMChannelDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_gcm_channel" {
				continue
			}

			_, err := tfpinpoint.FindGCMChannelByApplicationId(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Pinpoint GCM Channel %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGCMChannelConfig_basicAPIKey(apiKey string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_gcm_channel" "test_gcm_channel" {
  application_id = aws_pinpoint_app.test_app.application_id
  enabled        = "false"
  api_key        = "%s"
}
`, apiKey)
}

func testAccGCMChannelConfig_apiKeyAuthMethod(apiKey string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_gcm_channel" "test_gcm_channel" {
  application_id                = aws_pinpoint_app.test_app.application_id
  enabled                       = "false"
  default_authentication_method = "KEY"
  api_key                       = "%s"
}
`, apiKey)
}

func testAccGCMChannelConfig_tokenAuthMethod(serviceJsonFile string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_gcm_channel" "test_gcm_channel" {
  application_id                = aws_pinpoint_app.test_app.application_id
  enabled                       = "false"
  default_authentication_method = "TOKEN"
  service_json                  = file("%s")
}
`, serviceJsonFile)
}
