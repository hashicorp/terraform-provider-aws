// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfchime "github.com/hashicorp/terraform-provider-aws/internal/service/chime"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVoiceConnectorTerminationCredentials_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchime.ResourceVoiceConnectorTerminationCredentials(), resourceName),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccVoiceConnectorTerminationCredentials_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", acctest.Ct1),
				),
			},
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", acctest.Ct2),
				),
			},
		},
	})
}

func testAccCheckVoiceConnectorTerminationCredentialsExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime Voice Connector termination credentials ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

		_, err := tfchime.FindVoiceConnectorResourceWithRetry(ctx, false, func() (*chimesdkvoice.ListVoiceConnectorTerminationCredentialsOutput, error) {
			return tfchime.FindVoiceConnectorTerminationCredentialsByID(ctx, conn, rs.Primary.ID)
		})

		return err
	}
}

func testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chime_voice_connector_termination_credentials" {
				continue
			}
			conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

			_, err := tfchime.FindVoiceConnectorResourceWithRetry(ctx, false, func() (*chimesdkvoice.ListVoiceConnectorTerminationCredentialsOutput, error) {
				return tfchime.FindVoiceConnectorTerminationCredentialsByID(ctx, conn, rs.Primary.ID)
			})

			if tfresource.NotFound(err) {
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
