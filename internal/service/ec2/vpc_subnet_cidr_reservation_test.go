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

func TestAccVPCSubnetCIDRReservation_basic(t *testing.T) {
	var res ec2.SubnetCidrReservation
	resourceName := "aws_ec2_subnet_cidr_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetCIDRReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testSubnetCIDRReservationConfig_ipv4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetCIDRReservationExists(resourceName, &res),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.1.16/28"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "reservation_type", "prefix"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubnetCIDRReservationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSubnetCIDRReservation_ipv6(t *testing.T) {
	var res ec2.SubnetCidrReservation
	resourceName := "aws_ec2_subnet_cidr_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetCIDRReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testSubnetCIDRReservationConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetCIDRReservationExists(resourceName, &res),
					resource.TestCheckResourceAttr(resourceName, "reservation_type", "explicit"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubnetCIDRReservationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSubnetCIDRReservation_disappears(t *testing.T) {
	var res ec2.SubnetCidrReservation
	resourceName := "aws_ec2_subnet_cidr_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetCIDRReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testSubnetCIDRReservationConfig_ipv4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetCIDRReservationExists(resourceName, &res),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceSubnetCIDRReservation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSubnetCIDRReservationExists(n string, v *ec2.SubnetCidrReservation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Subnet CIDR Reservation ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindSubnetCIDRReservationBySubnetIDAndReservationID(conn, rs.Primary.Attributes["subnet_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSubnetCIDRReservationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_subnet_cidr_reservation" {
			continue
		}

		_, err := tfec2.FindSubnetCIDRReservationBySubnetIDAndReservationID(conn, rs.Primary.Attributes["subnet_id"], rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Subnet CIDR Reservation %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccSubnetCIDRReservationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		subnetId := rs.Primary.Attributes["subnet_id"]
		return fmt.Sprintf("%s:%s", subnetId, rs.Primary.ID), nil
	}
}

func testSubnetCIDRReservationConfig_ipv4(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_subnet_cidr_reservation" "test" {
  cidr_block       = "10.1.1.16/28"
  description      = "test"
  reservation_type = "prefix"
  subnet_id        = aws_subnet.test.id
}
`, rName)
}

func testSubnetCIDRReservationConfig_ipv6(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block      = "10.1.1.0/24"
  vpc_id          = aws_vpc.test.id
  ipv6_cidr_block = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_subnet_cidr_reservation" "test" {
  cidr_block       = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 12, 17)
  reservation_type = "explicit"
  subnet_id        = aws_subnet.test.id
}
`, rName)
}
