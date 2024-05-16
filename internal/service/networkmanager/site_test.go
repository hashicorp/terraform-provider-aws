// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerSite_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSiteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerSite_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkmanager.ResourceSite(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerSite_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSiteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSiteConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccNetworkManagerSite_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSiteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccNetworkManagerSite_location(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_location(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "location.0.address", "Stuart, FL"),
					resource.TestCheckResourceAttr(resourceName, "location.0.latitude", "27.198"),
					resource.TestCheckResourceAttr(resourceName, "location.0.longitude", "-80.253"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSiteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteConfig_locationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "location.0.address", "Brisbane, QLD"),
					resource.TestCheckResourceAttr(resourceName, "location.0.latitude", "-27.470"),
					resource.TestCheckResourceAttr(resourceName, "location.0.longitude", "153.026"),
				),
			},
		},
	})
}

func testAccCheckSiteDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_site" {
				continue
			}

			_, err := tfnetworkmanager.FindSiteByTwoPartKey(ctx, conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckSiteExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Site ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

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

func testAccSiteConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccSiteConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
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

func testAccSiteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrARN], nil
	}
}
