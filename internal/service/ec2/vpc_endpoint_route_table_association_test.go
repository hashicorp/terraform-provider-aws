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

func TestAccVPCEndpointRouteTableAssociation_basic(t *testing.T) {
	resourceName := "aws_vpc_endpoint_route_table_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointRouteTableAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointRouteTableAssociationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVPCEndpointRouteTableAssociationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCEndpointRouteTableAssociation_disappears(t *testing.T) {
	resourceName := "aws_vpc_endpoint_route_table_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCEndpointRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointRouteTableAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointRouteTableAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPCEndpointRouteTableAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCEndpointRouteTableAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint_route_table_association" {
			continue
		}

		err := tfec2.FindVPCEndpointRouteTableAssociationExists(conn, rs.Primary.Attributes["vpc_endpoint_id"], rs.Primary.Attributes["route_table_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("VPC Endpoint Route Table Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVPCEndpointRouteTableAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Endpoint Route Table Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		return tfec2.FindVPCEndpointRouteTableAssociationExists(conn, rs.Primary.Attributes["vpc_endpoint_id"], rs.Primary.Attributes["route_table_id"])
	}
}

func testAccVPCEndpointRouteTableAssociationImportStateIdFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		id := fmt.Sprintf("%s/%s", rs.Primary.Attributes["vpc_endpoint_id"], rs.Primary.Attributes["route_table_id"])
		return id, nil
	}
}

func testAccVPCEndpointRouteTableAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_route_table_association" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.test.id
  route_table_id  = aws_route_table.test.id
}
`, rName)
}
