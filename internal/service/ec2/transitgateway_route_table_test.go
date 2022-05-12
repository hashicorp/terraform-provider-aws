package ec2_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func testAccTransitGatewayRouteTable_basic(t *testing.T) {
	var transitGatewayRouteTable1 ec2.TransitGatewayRouteTable
	resourceName := "aws_ec2_transit_gateway_route_table.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableExists(resourceName, &transitGatewayRouteTable1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`transit-gateway-route-table/tgw-rtb-.+`)),
					resource.TestCheckResourceAttr(resourceName, "default_association_route_table", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_propagation_route_table", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccTransitGatewayRouteTable_disappears(t *testing.T) {
	var transitGatewayRouteTable1 ec2.TransitGatewayRouteTable
	resourceName := "aws_ec2_transit_gateway_route_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableExists(resourceName, &transitGatewayRouteTable1),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTransitGatewayRouteTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayRouteTable_disappears_TransitGateway(t *testing.T) {
	var transitGateway1 ec2.TransitGateway
	var transitGatewayRouteTable1 ec2.TransitGatewayRouteTable
	resourceName := "aws_ec2_transit_gateway_route_table.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckTransitGatewayRouteTableExists(resourceName, &transitGatewayRouteTable1),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTransitGateway(), transitGatewayResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayRouteTable_Tags(t *testing.T) {
	var transitGatewayRouteTable1, transitGatewayRouteTable2, transitGatewayRouteTable3 ec2.TransitGatewayRouteTable
	resourceName := "aws_ec2_transit_gateway_route_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableExists(resourceName, &transitGatewayRouteTable1),
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
				Config: testAccTransitGatewayRouteTableTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableExists(resourceName, &transitGatewayRouteTable2),
					testAccCheckTransitGatewayRouteTableNotRecreated(&transitGatewayRouteTable1, &transitGatewayRouteTable2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTransitGatewayRouteTableTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableExists(resourceName, &transitGatewayRouteTable3),
					testAccCheckTransitGatewayRouteTableNotRecreated(&transitGatewayRouteTable2, &transitGatewayRouteTable3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayRouteTableExists(resourceName string, transitGatewayRouteTable *ec2.TransitGatewayRouteTable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Route Table ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		routeTable, err := tfec2.DescribeTransitGatewayRouteTable(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if routeTable == nil {
			return fmt.Errorf("EC2 Transit Gateway Route Table not found")
		}

		if aws.StringValue(routeTable.State) != ec2.TransitGatewayRouteTableStateAvailable {
			return fmt.Errorf("EC2 Transit Gateway Route Table (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(routeTable.State))
		}

		*transitGatewayRouteTable = *routeTable

		return nil
	}
}

func testAccCheckTransitGatewayRouteTableDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route_table" {
			continue
		}

		routeTable, err := tfec2.DescribeTransitGatewayRouteTable(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
			continue
		}

		if err != nil {
			return err
		}

		if routeTable == nil {
			continue
		}

		if aws.StringValue(routeTable.State) != ec2.TransitGatewayRouteTableStateDeleted {
			return fmt.Errorf("EC2 Transit Gateway Route Table (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(routeTable.State))
		}
	}

	return nil
}

func testAccCheckTransitGatewayRouteTableNotRecreated(i, j *ec2.TransitGatewayRouteTable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayRouteTableId) != aws.StringValue(j.TransitGatewayRouteTableId) {
			return errors.New("EC2 Transit Gateway Route Table was recreated")
		}

		return nil
	}
}

func testAccTransitGatewayRouteTableConfig() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}
`
}

func testAccTransitGatewayRouteTableTags1Config(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1)
}

func testAccTransitGatewayRouteTableTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    %q = %q
    %q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
