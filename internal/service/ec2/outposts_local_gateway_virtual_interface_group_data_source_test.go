package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^lgw-vif-grp-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupDataSource_localGatewayID(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^lgw-vif-grp-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.source"
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", sourceDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "local_gateway_id", sourceDataSourceName, "local_gateway_id"),
				),
			},
		},
	})
}

func testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_filter() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}
`
}

func testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_id() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}
`
}

func testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "source" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}

resource "aws_ec2_tag" "test" {
  key         = "TerraformAccTest-aws_ec2_local_gateway_virtual_interface_group"
  resource_id = data.aws_ec2_local_gateway_virtual_interface_group.source.id
  value       = %[1]q
}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  tags = {
    (aws_ec2_tag.test.key) = aws_ec2_tag.test.value
  }
}
`, rName)
}
