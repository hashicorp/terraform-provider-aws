package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2NetworkInterfacesDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
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

func TestAccEC2NetworkInterfacesDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
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

func testAccNetworkInterfacesDataSourceConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-eni-data-source-basic-%s"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-eni-data-source-basic-%s"
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_network_interface" "test1" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = aws_vpc.test.tags.Name
  }
}
`, rName, rName)
}

func testAccNetworkInterfacesDataSourceConfig_Filter(rName string) string {
	return testAccNetworkInterfacesDataSourceConfig_Base(rName) + `
data "aws_network_interfaces" "test" {
  filter {
    name   = "subnet-id"
    values = [aws_network_interface.test.subnet_id, aws_network_interface.test1.subnet_id]
  }
}
`
}

func testAccNetworkInterfacesDataSourceConfig_Tags(rName string) string {
	return testAccNetworkInterfacesDataSourceConfig_Base(rName) + `
data "aws_network_interfaces" "test" {
  tags = {
    Name = aws_network_interface.test1.tags.Name
  }
}
`
}
