package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayVirtualInterfaceGroupsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupsDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayVirtualInterfaceGroupsFilterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupsDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayVirtualInterfaceGroupsTagsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func testAccLocalGatewayVirtualInterfaceGroupsDataSourceConfig() string {
	return `
data "aws_ec2_local_gateway_virtual_interface_groups" "test" {}
`
}

func testAccLocalGatewayVirtualInterfaceGroupsFilterDataSourceConfig() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_groups" "test" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}
`
}

func testAccLocalGatewayVirtualInterfaceGroupsTagsDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_groups" "source" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}

resource "aws_ec2_tag" "test" {
  key         = "TerraformAccTest-aws_ec2_local_gateway_virtual_interface_groups"
  resource_id = tolist(data.aws_ec2_local_gateway_virtual_interface_groups.source.ids)[0]
  value       = %[1]q
}

data "aws_ec2_local_gateway_virtual_interface_groups" "test" {
  tags = {
    (aws_ec2_tag.test.key) = aws_ec2_tag.test.value
  }
}
`, rName)
}
