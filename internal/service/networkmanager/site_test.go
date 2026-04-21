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

func TestAccNetworkManagerSite_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_site.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, t, resourceName),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "networkmanager", "site/{global_network_id}/{id}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerSite_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_site.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfnetworkmanager.ResourceSite(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerSite_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_site.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccNetworkManagerSite_location(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_site.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_location(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.address", "Stuart, FL"),
					resource.TestCheckResourceAttr(resourceName, "location.0.latitude", "27.198"),
					resource.TestCheckResourceAttr(resourceName, "location.0.longitude", "-80.253"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteConfig_locationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.address", "Brisbane, QLD"),
					resource.TestCheckResourceAttr(resourceName, "location.0.latitude", "-27.470"),
					resource.TestCheckResourceAttr(resourceName, "location.0.longitude", "153.026"),
				),
			},
		},
	})
}

func testAccCheckSiteDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_site" {
				continue
			}

			_, err := tfnetworkmanager.FindSiteByTwoPartKey(ctx, conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Site %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSiteExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Site ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		_, err := tfnetworkmanager.FindSiteByTwoPartKey(ctx, conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

		return err
	}
}

func testAccSiteConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}
`, rName)
}

func testAccSiteConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  description       = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, description)
}

func testAccSiteConfig_location(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  location {
    address   = "Stuart, FL"
    latitude  = "27.198"
    longitude = "-80.253"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccSiteConfig_locationUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  location {
    address   = "Brisbane, QLD"
    latitude  = "-27.470"
    longitude = "153.026"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
