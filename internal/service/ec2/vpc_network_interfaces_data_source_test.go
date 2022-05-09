package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCNetworkInterfacesDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfacesDataSourceConfig_Filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_network_interfaces.test", "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterfacesDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfacesDataSourceConfig_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_network_interfaces.test", "ids.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterfacesDataSource_empty(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfacesDataSourceConfig_Empty(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_network_interfaces.test", "ids.#", "0"),
				),
			},
		},
	})
}

func testAccNetworkInterfacesDataSourceConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test1" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = "%[1]s-1"
  }
}

resource "aws_network_interface" "test2" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = "%[1]s-2"
  }
}
`, rName)
}

func testAccNetworkInterfacesDataSourceConfig_Filter(rName string) string {
	return acctest.ConfigCompose(testAccNetworkInterfacesDataSourceConfig_Base(rName), `
data "aws_network_interfaces" "test" {
  filter {
    name   = "subnet-id"
    values = [aws_network_interface.test1.subnet_id, aws_network_interface.test2.subnet_id]
  }
}
`)
}

func testAccNetworkInterfacesDataSourceConfig_Tags(rName string) string {
	return acctest.ConfigCompose(testAccNetworkInterfacesDataSourceConfig_Base(rName), `
data "aws_network_interfaces" "test" {
  tags = {
    Name = aws_network_interface.test2.tags.Name
  }
}
`)
}

func testAccNetworkInterfacesDataSourceConfig_Empty(rName string) string {
	return fmt.Sprintf(`
data "aws_network_interfaces" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
