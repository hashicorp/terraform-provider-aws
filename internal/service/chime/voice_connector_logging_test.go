// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package chime_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfchime "github.com/hashicorp/terraform-provider-aws/internal/service/chime"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVoiceConnectorLogging_basic(t *testing.T) {
	ctx := acctest.Context(t)
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_logging.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorLoggingConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorLoggingExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_sip_logs", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_media_metric_logs", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVoiceConnectorLogging_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_logging.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorLoggingConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceConnectorLoggingExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfchime.ResourceVoiceConnectorLogging(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVoiceConnectorLogging_update(t *testing.T) {
	ctx := acctest.Context(t)
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_logging.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorLoggingConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorLoggingExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccVoiceConnectorLoggingConfig_updated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorLoggingExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_sip_logs", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_media_metric_logs", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVoiceConnectorLoggingConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_logging" "test" {
  voice_connector_id       = aws_chime_voice_connector.chime.id
  enable_sip_logs          = true
  enable_media_metric_logs = true
}
`, name)
}

func testAccVoiceConnectorLoggingConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_logging" "test" {
  voice_connector_id       = aws_chime_voice_connector.chime.id
  enable_sip_logs          = false
  enable_media_metric_logs = false
}
`, name)
}

func testAccCheckVoiceConnectorLoggingExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime Voice Connector logging ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ChimeSDKVoiceClient(ctx)

		_, err := tfchime.FindVoiceConnectorResourceWithRetry(ctx, false, func() (*awstypes.LoggingConfiguration, error) {
			return tfchime.FindVoiceConnectorLoggingByID(ctx, conn, rs.Primary.ID)
		})

		return err
	}
}
