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

func TestAccNetworkManagerConnection_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":               testAccConnection_basic,
		"disappears":          testAccConnection_disappears,
		"tags":                testAccConnection_tags,
		"descriptionAndLinks": testAccConnection_descriptionAndLinks,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccConnection_basic(t *testing.T) {
	resourceName := "aws_networkmanager_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "connected_link_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "link_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccConnectionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccConnection_disappears(t *testing.T) {
	resourceName := "aws_networkmanager_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfnetworkmanager.ResourceConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConnection_tags(t *testing.T) {
	resourceName := "aws_networkmanager_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccConnectionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccConnectionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccConnectionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccConnection_descriptionAndLinks(t *testing.T) {
	resourceName := "aws_networkmanager_connection.test"
	link1ResourceName := "aws_networkmanager_link.test1"
	link2ResourceName := "aws_networkmanager_link.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDescriptionAndLinksConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "connected_link_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttrPair(resourceName, "link_id", link1ResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccConnectionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccConnectionDescriptionAndLinksUpdatedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "connected_link_id", link2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttrPair(resourceName, "link_id", link1ResourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckConnectionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_device" {
			continue
		}

		_, err := tfnetworkmanager.FindConnectionByTwoPartKey(context.TODO(), conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Network Manager Connection %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckConnectionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Connection ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

		_, err := tfnetworkmanager.FindConnectionByTwoPartKey(context.TODO(), conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccConnectionBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  description = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  description = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id

  description = "%[1]s-1"

  site_id = aws_networkmanager_site.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id

  description = "%[1]s-2"

  site_id = aws_networkmanager_site.test.id

  tags = {
    Name = %[1]q
  }

  # Create one device at a time.
  depends_on = [aws_networkmanager_device.test1]
}
`, rName)
}

func testAccConnectionConfig(rName string) string {
	return acctest.ConfigCompose(testAccConnectionBaseConfig(rName), `
resource "aws_networkmanager_connection" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  device_id           = aws_networkmanager_device.test1.id
  connected_device_id = aws_networkmanager_device.test2.id
}
`)
}

func testAccConnectionConfigTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccConnectionBaseConfig(rName), fmt.Sprintf(`
resource "aws_networkmanager_connection" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  device_id           = aws_networkmanager_device.test1.id
  connected_device_id = aws_networkmanager_device.test2.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccConnectionConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccConnectionBaseConfig(rName), fmt.Sprintf(`
resource "aws_networkmanager_connection" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  device_id           = aws_networkmanager_device.test1.id
  connected_device_id = aws_networkmanager_device.test2.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccConnectionDescriptionAndLinksBaseConfig(rName string) string {
	return acctest.ConfigCompose(testAccConnectionBaseConfig(rName), fmt.Sprintf(`
resource "aws_networkmanager_link" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  description = "%[1]s-1"

  bandwidth {
    download_speed = 50
    upload_speed   = 10
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_link" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  description = "%[1]s-2"

  bandwidth {
    download_speed = 100
    upload_speed   = 20
  }

  tags = {
    Name = %[1]q
  }

  # Create one link at a time.
  depends_on = [aws_networkmanager_link.test1]
}

resource "aws_networkmanager_link_association" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id
  link_id           = aws_networkmanager_link.test1.id
  device_id         = aws_networkmanager_device.test1.id
}

resource "aws_networkmanager_link_association" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id
  link_id           = aws_networkmanager_link.test2.id
  device_id         = aws_networkmanager_device.test2.id
}
`, rName))
}

func testAccConnectionDescriptionAndLinksConfig(rName string) string {
	return acctest.ConfigCompose(testAccConnectionDescriptionAndLinksBaseConfig(rName), fmt.Sprintf(`
resource "aws_networkmanager_connection" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  device_id           = aws_networkmanager_device.test1.id
  connected_device_id = aws_networkmanager_device.test2.id

  description = "description1"

  link_id = aws_networkmanager_link.test1.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_networkmanager_link_association.test1, aws_networkmanager_link_association.test2]
}
`, rName))
}

func testAccConnectionDescriptionAndLinksUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(testAccConnectionDescriptionAndLinksBaseConfig(rName), fmt.Sprintf(`
resource "aws_networkmanager_connection" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  device_id           = aws_networkmanager_device.test1.id
  connected_device_id = aws_networkmanager_device.test2.id

  description = "description2"

  link_id           = aws_networkmanager_link.test1.id
  connected_link_id = aws_networkmanager_link.test2.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_networkmanager_link_association.test1, aws_networkmanager_link_association.test2]
}
`, rName))
}

func testAccConnectionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["arn"], nil
	}
}
