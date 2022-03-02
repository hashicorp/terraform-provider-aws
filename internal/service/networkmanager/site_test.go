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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, networkmanager.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSiteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfig(rName),
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

/*
func TestAccSite_tags(t *testing.T) {
	resourceName := "aws_networkmanager_site.test"
	description := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSiteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteConfigTags1(description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
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
				Config: testAccSiteConfigTags2(description, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSiteConfigTags1(description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSiteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}
*/

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

func testAccSiteConfig(rName string) string {
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

func testAccSiteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["arn"], nil
	}
}
