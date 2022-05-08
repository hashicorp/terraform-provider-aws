package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2OutpostsLocalGatewayRouteTableVPCAssociation_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	localGatewayRouteTableDataSourceName := "data.aws_ec2_local_gateway_route_table.test"
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocalGatewayRouteTableVPCAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteTableVPCAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVPCAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", localGatewayRouteTableDataSourceName, "local_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_route_table_id", localGatewayRouteTableDataSourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayRouteTableVPCAssociation_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocalGatewayRouteTableVPCAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteTableVPCAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVPCAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceLocalGatewayRouteTableVPCAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayRouteTableVPCAssociation_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocalGatewayRouteTableVPCAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteTableVPCAssociationTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVPCAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLocalGatewayRouteTableVPCAssociationTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVPCAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocalGatewayRouteTableVPCAssociationTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVPCAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckLocalGatewayRouteTableVPCAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("%s: missing resource ID", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		association, err := tfec2.GetLocalGatewayRouteTableVPCAssociation(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if association == nil {
			return fmt.Errorf("EC2 Local Gateway Route Table VPC Association (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(association.State) != ec2.RouteTableAssociationStateCodeAssociated {
			return fmt.Errorf("EC2 Local Gateway Route Table VPC Association (%s) not in associated state: %s", rs.Primary.ID, aws.StringValue(association.State))
		}

		return nil
	}
}

func testAccCheckLocalGatewayRouteTableVPCAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_local_gateway_route_table_vpc_association" {
			continue
		}

		association, err := tfec2.GetLocalGatewayRouteTableVPCAssociation(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if association != nil && aws.StringValue(association.State) != ec2.RouteTableAssociationStateCodeDisassociated {
			return fmt.Errorf("EC2 Local Gateway Route Table VPC Association (%s) still exists in state: %s", rs.Primary.ID, aws.StringValue(association.State))
		}
	}

	return nil
}

func testAccLocalGatewayRouteTableVPCAssociationBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccLocalGatewayRouteTableVPCAssociationConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableVPCAssociationBaseConfig(rName),
		`
resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id
}
`)
}

func testAccLocalGatewayRouteTableVPCAssociationTags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableVPCAssociationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccLocalGatewayRouteTableVPCAssociationTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableVPCAssociationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
