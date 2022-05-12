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

func TestAccNetworkManagerDevice_basic(t *testing.T) {
	resourceName := "aws_networkmanager_device.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "aws_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "model", ""),
					resource.TestCheckResourceAttr(resourceName, "serial_number", ""),
					resource.TestCheckResourceAttr(resourceName, "site_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", ""),
					resource.TestCheckResourceAttr(resourceName, "vendor", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeviceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerDevice_disappears(t *testing.T) {
	resourceName := "aws_networkmanager_device.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfnetworkmanager.ResourceDevice(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerDevice_tags(t *testing.T) {
	resourceName := "aws_networkmanager_device.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeviceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDeviceConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccNetworkManagerDevice_allAttributes(t *testing.T) {
	resourceName := "aws_networkmanager_device.test"
	site1ResourceName := "aws_networkmanager_site.test1"
	site2ResourceName := "aws_networkmanager_site.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceAllAttributesConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.address", "Address 1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.latitude", "1.1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.longitude", "-1.1"),
					resource.TestCheckResourceAttr(resourceName, "model", "model1"),
					resource.TestCheckResourceAttr(resourceName, "serial_number", "sn1"),
					resource.TestCheckResourceAttrPair(resourceName, "site_id", site1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "type1"),
					resource.TestCheckResourceAttr(resourceName, "vendor", "vendor1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeviceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceAllAttributesUpdatedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.address", "Address 2"),
					resource.TestCheckResourceAttr(resourceName, "location.0.latitude", "22"),
					resource.TestCheckResourceAttr(resourceName, "location.0.longitude", "-22"),
					resource.TestCheckResourceAttr(resourceName, "model", "model2"),
					resource.TestCheckResourceAttr(resourceName, "serial_number", "sn2"),
					resource.TestCheckResourceAttrPair(resourceName, "site_id", site2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "type2"),
					resource.TestCheckResourceAttr(resourceName, "vendor", "vendor2"),
				),
			},
		},
	})
}

func TestAccNetworkManagerDevice_awsLocation(t *testing.T) {
	resourceName := "aws_networkmanager_device.test"
	subnetResourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceAWSLocationConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aws_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "aws_location.0.subnet_arn", subnetResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "aws_location.0.zone", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeviceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceAWSLocationUpdatedConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aws_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "aws_location.0.subnet_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "aws_location.0.zone", subnetResourceName, "availability_zone"),
				),
			},
		},
	})
}

func testAccCheckDeviceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_device" {
			continue
		}

		_, err := tfnetworkmanager.FindDeviceByTwoPartKey(context.TODO(), conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Network Manager Device %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckDeviceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Device ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

		_, err := tfnetworkmanager.FindDeviceByTwoPartKey(context.TODO(), conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccDeviceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}
`, rName)
}

func testAccDeviceConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDeviceConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDeviceAllAttributesConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  description   = "description1"
  model         = "model1"
  serial_number = "sn1"
  site_id       = aws_networkmanager_site.test1.id
  type          = "type1"
  vendor        = "vendor1"

  location {
    address   = "Address 1"
    latitude  = "1.1"
    longitude = "-1.1"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccDeviceAllAttributesUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  description   = "description2"
  model         = "model2"
  serial_number = "sn2"
  site_id       = aws_networkmanager_site.test2.id
  type          = "type2"
  vendor        = "vendor2"

  location {
    address   = "Address 2"
    latitude  = "22"
    longitude = "-22"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccDeviceAWSLocationConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  aws_location {
    subnet_arn = aws_subnet.test.arn
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccDeviceAWSLocationUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  aws_location {
    zone = aws_subnet.test.availability_zone
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccDeviceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["arn"], nil
	}
}
