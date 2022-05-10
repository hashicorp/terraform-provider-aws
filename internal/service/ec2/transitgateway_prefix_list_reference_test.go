package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccTransitGatewayPrefixListReference_basic(t *testing.T) {
	managedPrefixListResourceName := "aws_ec2_managed_prefix_list.test"
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			testAccPreCheckEc2ManagedPrefixList(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayPrefixListReferenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPrefixListReferenceConfig_Blackhole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccTransitGatewayPrefixListReferenceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "blackhole", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", managedPrefixListResourceName, "id"),
					acctest.CheckResourceAttrAccountID(resourceName, "prefix_list_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_attachment_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayResourceName, "association_default_route_table_id"),
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

func testAccTransitGatewayPrefixListReference_disappears(t *testing.T) {
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			testAccPreCheckEc2ManagedPrefixList(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayPrefixListReferenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPrefixListReferenceConfig_Blackhole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccTransitGatewayPrefixListReferenceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTransitGatewayPrefixListReference(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayPrefixListReference_disappears_TransitGateway(t *testing.T) {
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			testAccPreCheckEc2ManagedPrefixList(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayPrefixListReferenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPrefixListReferenceConfig_Blackhole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccTransitGatewayPrefixListReferenceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTransitGateway(), transitGatewayResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayPrefixListReference_TransitGatewayAttachmentID(t *testing.T) {
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	transitGatewayVpcAttachmentResourceName1 := "aws_ec2_transit_gateway_vpc_attachment.test.0"
	transitGatewayVpcAttachmentResourceName2 := "aws_ec2_transit_gateway_vpc_attachment.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			testAccPreCheckEc2ManagedPrefixList(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayPrefixListReferenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPrefixListReferenceConfig_TransitGatewayAttachmentID(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccTransitGatewayPrefixListReferenceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "blackhole", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayVpcAttachmentResourceName1, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayPrefixListReferenceConfig_TransitGatewayAttachmentID(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccTransitGatewayPrefixListReferenceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "blackhole", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayVpcAttachmentResourceName2, "id"),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayPrefixListReferenceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_prefix_list_reference" {
			continue
		}

		transitGatewayRouteTableID, prefixListID, err := tfec2.TransitGatewayPrefixListReferenceParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfec2.FindTransitGatewayPrefixListReferenceByTwoPartKey(conn, transitGatewayRouteTableID, prefixListID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Transit Gateway Prefix List Reference %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccTransitGatewayPrefixListReferenceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Prefix List Reference is set")
		}

		transitGatewayRouteTableID, prefixListID, err := tfec2.TransitGatewayPrefixListReferenceParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		_, err = tfec2.FindTransitGatewayPrefixListReferenceByTwoPartKey(conn, transitGatewayRouteTableID, prefixListID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccTransitGatewayPrefixListReferenceConfig_Blackhole(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_prefix_list_reference" "test" {
  blackhole                      = true
  prefix_list_id                 = aws_ec2_managed_prefix_list.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}
`, rName)
}

func testAccTransitGatewayPrefixListReferenceConfig_TransitGatewayAttachmentID(rName string, index int) string {
	return fmt.Sprintf(`
variable "index" {
  default = %[2]d
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_vpc" "test" {
  count = 2

  cidr_block = "10.${count.index}.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  cidr_block = cidrsubnet(aws_vpc.test[count.index].cidr_block, 8, 0)
  vpc_id     = aws_vpc.test[count.index].id
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  count = 2

  subnet_ids         = [aws_subnet.test[count.index].id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test[count.index].id
}

resource "aws_ec2_transit_gateway_prefix_list_reference" "test" {
  prefix_list_id                 = aws_ec2_managed_prefix_list.test.id
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test[var.index].id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}
`, rName, index)
}
