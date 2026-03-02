// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerLinkAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_link_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLinkAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkAssociationExists(ctx, t, resourceName),
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

func TestAccNetworkManagerLinkAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_link_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLinkAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfnetworkmanager.ResourceLinkAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLinkAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_link_association" {
				continue
			}

			globalNetworkID, linkID, deviceID, err := tfnetworkmanager.LinkAssociationParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfnetworkmanager.FindLinkAssociationByThreePartKey(ctx, conn, globalNetworkID, linkID, deviceID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Link Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLinkAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Link Association ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		globalNetworkID, linkID, deviceID, err := tfnetworkmanager.LinkAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfnetworkmanager.FindLinkAssociationByThreePartKey(ctx, conn, globalNetworkID, linkID, deviceID)

		return err
	}
}

func testAccLinkAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_link" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  bandwidth {
    download_speed = 50
    upload_speed   = 10
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_link_association" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  link_id           = aws_networkmanager_link.test.id
  device_id         = aws_networkmanager_device.test.id
}
`, rName)
}
