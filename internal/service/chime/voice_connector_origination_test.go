// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfchime "github.com/hashicorp/terraform-provider-aws/internal/service/chime"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVoiceConnectorOrigination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_origination.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorOriginationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorOriginationConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorOriginationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						names.AttrProtocol: "TCP",
						names.AttrPriority: acctest.Ct1,
					}),
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

func testAccVoiceConnectorOrigination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_origination.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorOriginationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorOriginationConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceConnectorOriginationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchime.ResourceVoiceConnectorOrigination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVoiceConnectorOrigination_update(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_origination.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorOriginationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorOriginationConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorOriginationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
				),
			},
			{
				Config: testAccVoiceConnectorOriginationConfig_updated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorOriginationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						names.AttrProtocol: "TCP",
						names.AttrPort:     "5060",
						names.AttrPriority: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						names.AttrProtocol: "UDP",
						names.AttrPriority: acctest.Ct2,
					}),
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

func testAccCheckVoiceConnectorOriginationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime voice connector origination ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

		_, err := tfchime.FindVoiceConnectorResourceWithRetry(ctx, false, func() (*awstypes.Origination, error) {
			return tfchime.FindVoiceConnectorOriginationByID(ctx, conn, rs.Primary.ID)
		})

		return err
	}
}

func testAccCheckVoiceConnectorOriginationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chime_voice_connector_origination" {
				continue
			}
			conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

			_, err := tfchime.FindVoiceConnectorResourceWithRetry(ctx, false, func() (*awstypes.Origination, error) {
				return tfchime.FindVoiceConnectorOriginationByID(ctx, conn, rs.Primary.ID)
			})

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("voice connector origination still exists: (%s)", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVoiceConnectorOriginationConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "test" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_origination" "test" {
  route {
    host     = "200.100.12.1"
    port     = 5060
    protocol = "TCP"
    priority = 1
    weight   = 1
  }
  voice_connector_id = aws_chime_voice_connector.test.id
}
`, name)
}

func testAccVoiceConnectorOriginationConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "test" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_origination" "test" {
  voice_connector_id = aws_chime_voice_connector.test.id

  route {
    host     = "200.100.12.1"
    port     = 5060
    protocol = "TCP"
    priority = 1
    weight   = 1
  }

  route {
    host     = "209.166.124.147"
    protocol = "UDP"
    priority = 2
    weight   = 30
  }
}
`, name)
}
