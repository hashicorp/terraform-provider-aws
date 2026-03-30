// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package chime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfchime "github.com/hashicorp/terraform-provider-aws/internal/service/chime"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVoiceConnectorTerminationCredentials_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credentials"},
			},
		},
	})
}

func testAccVoiceConnectorTerminationCredentials_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfchime.ResourceVoiceConnectorTerminationCredentials(), resourceName),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccVoiceConnectorTerminationCredentials_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "1"),
				),
			},
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "2"),
				),
			},
		},
	})
}

func testAccCheckVoiceConnectorTerminationCredentialsExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime Voice Connector termination credentials ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ChimeSDKVoiceClient(ctx)

		_, err := tfchime.FindVoiceConnectorResourceWithRetry(ctx, false, func() (*chimesdkvoice.ListVoiceConnectorTerminationCredentialsOutput, error) {
			return tfchime.FindVoiceConnectorTerminationCredentialsByID(ctx, conn, rs.Primary.ID)
		})

		return err
	}
}

func testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chime_voice_connector_termination_credentials" {
				continue
			}
			conn := acctest.ProviderMeta(ctx, t).ChimeSDKVoiceClient(ctx)

			_, err := tfchime.FindVoiceConnectorResourceWithRetry(ctx, false, func() (*chimesdkvoice.ListVoiceConnectorTerminationCredentialsOutput, error) {
				return tfchime.FindVoiceConnectorTerminationCredentialsByID(ctx, conn, rs.Primary.ID)
			})

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("voice connector terimination credentials still exists: (%s)", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVoiceConnectorTerminationCredentialsBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_termination" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  calling_regions = ["US"]
  cidr_allow_list = ["50.35.78.0/27"]
}
`, rName)
}

func testAccVoiceConnectorTerminationCredentialsConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVoiceConnectorTerminationCredentialsBaseConfig(rName), `
resource "aws_chime_voice_connector_termination_credentials" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  credentials {
    username = "test1"
    password = "test1!"
  }

  depends_on = [aws_chime_voice_connector_termination.test]
}
`)
}

func testAccVoiceConnectorTerminationCredentialsConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccVoiceConnectorTerminationCredentialsBaseConfig(rName), `
resource "aws_chime_voice_connector_termination_credentials" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  credentials {
    username = "test1"
    password = "test1!"
  }

  credentials {
    username = "test2"
    password = "test2!"
  }

  depends_on = [aws_chime_voice_connector_termination.test]
}
`)
}
