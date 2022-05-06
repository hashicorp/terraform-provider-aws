package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCPeeringConnectionsDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_peering_connections.test_by_filters", "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionsDataSource_NoMatches(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionsDataSourceConfig_NoMatches(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_peering_connections.test", "ids.#", "0"),
				),
			},
		},
	})
}

func testAccVPCPeeringConnectionsDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test3" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test1" {
  vpc_id      = aws_vpc.test1.id
  peer_vpc_id = aws_vpc.test2.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test2" {
  vpc_id      = aws_vpc.test1.id
  peer_vpc_id = aws_vpc.test3.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_peering_connections" "test_by_filters" {
  filter {
    name   = "vpc-peering-connection-id"
    values = [aws_vpc_peering_connection.test1.id, aws_vpc_peering_connection.test2.id]
  }
}
`, rName)
}

func testAccVPCPeeringConnectionsDataSourceConfig_NoMatches(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc_peering_connections" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
