package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccNetworkManagerSite_basic(t *testing.T) {
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSiteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSiteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfnetworkmanager.ResourceSite(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerSite_tags(t *testing.T) {
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSiteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSiteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSiteConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccNetworkManagerSite_description(t *testing.T) {
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSiteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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
					testAccCheckSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccNetworkManagerSite_location(t *testing.T) {
	resourceName := "aws_networkmanager_site.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSiteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig_location(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
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
					testAccCheckSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.address", "Brisbane, QLD"),
					resource.TestCheckResourceAttr(resourceName, "location.0.latitude", "-27.470"),
					resource.TestCheckResourceAttr(resourceName, "location.0.longitude", "153.026"),
				),
			},
		},
	})
}

func testAccCheckSiteDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_site" {
			continue
		}

		_, err := tfnetworkmanager.FindSiteByTwoPartKey(context.TODO(), conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

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

func testAccCheckSiteExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Site ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

		_, err := tfnetworkmanager.FindSiteByTwoPartKey(context.TODO(), conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
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

		return rs.Primary.Attributes["arn"], nil
	}
}
