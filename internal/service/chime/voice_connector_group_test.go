// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfchime "github.com/hashicorp/terraform-provider-aws/internal/service/chime"
)

func TestAccChimeVoiceConnectorGroup_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"basic":      testAccVoiceConnectorGroup_basic,
		"disappears": testAccVoiceConnectorGroup_disappears,
		"update":     testAccVoiceConnectorGroup_update,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccVoiceConnectorGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var voiceConnectorGroup *chimesdkvoice.VoiceConnectorGroup

	vcgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkvoice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorGroupConfig_basic(vcgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorGroupExists(ctx, resourceName, voiceConnectorGroup),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("vcg-%s", vcgName)),
					resource.TestCheckResourceAttr(resourceName, "connector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector.0.priority", "1"),
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

func testAccVoiceConnectorGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var voiceConnectorGroup *chimesdkvoice.VoiceConnectorGroup

	vcgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkvoice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorGroupConfig_basic(vcgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceConnectorGroupExists(ctx, resourceName, voiceConnectorGroup),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchime.ResourceVoiceConnectorGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVoiceConnectorGroup_update(t *testing.T) {
	ctx := acctest.Context(t)
	var voiceConnectorGroup *chimesdkvoice.VoiceConnectorGroup

	vcgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkvoice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorGroupConfig_basic(vcgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorGroupExists(ctx, resourceName, voiceConnectorGroup),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("vcg-%s", vcgName)),
					resource.TestCheckResourceAttr(resourceName, "connector.#", "1"),
				),
			},
			{
				Config: testAccVoiceConnectorGroupConfig_updated(vcgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("vcg-updated-%s", vcgName)),
					resource.TestCheckResourceAttr(resourceName, "connector.0.priority", "3"),
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

func testAccVoiceConnectorGroupConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_group" "test" {
  name = "vcg-%[1]s"

  connector {
    voice_connector_id = aws_chime_voice_connector.chime.id
    priority           = 1
  }
}
`, name)
}

func testAccVoiceConnectorGroupConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = false
}

resource "aws_chime_voice_connector_group" "test" {
  name = "vcg-updated-%[1]s"

  connector {
    voice_connector_id = aws_chime_voice_connector.chime.id
    priority           = 3
  }
}
`, name)
}

func testAccCheckVoiceConnectorGroupExists(ctx context.Context, name string, vc *chimesdkvoice.VoiceConnectorGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime voice connector group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceConn(ctx)
		input := &chimesdkvoice.GetVoiceConnectorGroupInput{
			VoiceConnectorGroupId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetVoiceConnectorGroupWithContext(ctx, input)
		if err != nil || resp.VoiceConnectorGroup == nil {
			return err
		}

		vc = resp.VoiceConnectorGroup
		return nil
	}
}

func testAccCheckVoiceConnectorGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chime_voice_connector" {
				continue
			}
			conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceConn(ctx)
			input := &chimesdkvoice.GetVoiceConnectorGroupInput{
				VoiceConnectorGroupId: aws.String(rs.Primary.ID),
			}
			resp, err := conn.GetVoiceConnectorGroupWithContext(ctx, input)
			if err == nil {
				if resp.VoiceConnectorGroup != nil && aws.StringValue(resp.VoiceConnectorGroup.Name) != "" {
					return fmt.Errorf("error Chime Voice Connector still exists")
				}
			}
			return nil
		}
		return nil
	}
}
