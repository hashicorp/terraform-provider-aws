package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCInternetGatewayDataSource_basic(t *testing.T) {
	igwResourceName := "aws_internet_gateway.test"
	vpcResourceName := "aws_vpc.test"
	ds1ResourceName := "data.aws_internet_gateway.by_id"
	ds2ResourceName := "data.aws_internet_gateway.by_filter"
	ds3ResourceName := "data.aws_internet_gateway.by_tags"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInternetGatewayDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "internet_gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", igwResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "attachments.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", igwResourceName, "arn"),

					resource.TestCheckResourceAttrPair(ds2ResourceName, "internet_gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "owner_id", igwResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "attachments.0.vpc_id", vpcResourceName, "id"),

					resource.TestCheckResourceAttrPair(ds3ResourceName, "internet_gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "owner_id", igwResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "attachments.0.vpc_id", vpcResourceName, "id"),
				),
			},
		},
	})
}

func testAccInternetGatewayDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_internet_gateway" "by_id" {
  internet_gateway_id = aws_internet_gateway.test.id
}

data "aws_internet_gateway" "by_tags" {
  tags = {
    Name = aws_internet_gateway.test.tags["Name"]
  }
}

data "aws_internet_gateway" "by_filter" {
  filter {
    name   = "internet-gateway-id"
    values = [aws_internet_gateway.test.id]
  }
}
`, rName)
}
