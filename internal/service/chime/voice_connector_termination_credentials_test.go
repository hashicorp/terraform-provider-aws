// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfchime "github.com/hashicorp/terraform-provider-aws/internal/service/chime"
)

func TestAccChimeVoiceConnectorTerminationCredentials_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, chime.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, resourceName),
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

func TestAccChimeVoiceConnectorTerminationCredentials_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, chime.EndpointsID),
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

func TestAccChimeVoiceConnectorTerminationCredentials_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, chime.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "1"),
				),
			},
			{
				Config: testAccVoiceConnectorTerminationCredentialsConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorTerminationCredentialsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "2"),
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn(ctx)
		input := &chime.ListVoiceConnectorTerminationCredentialsInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.ListVoiceConnectorTerminationCredentialsWithContext(ctx, input)
		if err != nil {
			return err
		}

		if resp == nil || resp.Usernames == nil {
			return fmt.Errorf("no Chime Voice Connector Termintation credentials (%s) found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVoiceConnectorTerminationCredentialsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chime_voice_connector_termination_credentials" {
				continue
			}
			conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn(ctx)
			input := &chime.ListVoiceConnectorTerminationCredentialsInput{
				VoiceConnectorId: aws.String(rs.Primary.ID),
			}
			resp, err := conn.ListVoiceConnectorTerminationCredentialsWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, chime.ErrCodeNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}

			if resp != nil && resp.Usernames != nil {
				return fmt.Errorf("error Chime Voice Connector Termination credentials still exists")
			}
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
