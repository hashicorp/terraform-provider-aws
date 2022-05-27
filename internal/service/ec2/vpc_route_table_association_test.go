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

func TestAccVPCRouteTableAssociation_Subnet_basic(t *testing.T) {
	var rta ec2.RouteTableAssociation
	resourceName := "aws_route_table_association.test"
	resourceNameRouteTable := "aws_route_table.test"
	resourceNameSubnet := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableAssociationConfig_subnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", resourceNameSubnet, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteTabAssocImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_Subnet_changeRouteTable(t *testing.T) {
	var rta ec2.RouteTableAssociation
	resourceName := "aws_route_table_association.test"
	resourceNameRouteTable1 := "aws_route_table.test"
	resourceNameRouteTable2 := "aws_route_table.test2"
	resourceNameSubnet := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableAssociationConfig_subnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable1, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", resourceNameSubnet, "id"),
				),
			},
			{
				Config: testAccVPCRouteTableAssociationConfig_subnetChange(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable2, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", resourceNameSubnet, "id"),
				),
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_Gateway_basic(t *testing.T) {
	var rta ec2.RouteTableAssociation
	resourceName := "aws_route_table_association.test"
	resourceNameRouteTable := "aws_route_table.test"
	resourceNameGateway := "aws_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableAssociationConfig_gateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", resourceNameGateway, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteTabAssocImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_Gateway_changeRouteTable(t *testing.T) {
	var rta ec2.RouteTableAssociation
	resourceName := "aws_route_table_association.test"
	resourceNameRouteTable1 := "aws_route_table.test"
	resourceNameRouteTable2 := "aws_route_table.test2"
	resourceNameGateway := "aws_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableAssociationConfig_gateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable1, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", resourceNameGateway, "id"),
				),
			},
			{
				Config: testAccVPCRouteTableAssociationConfig_gatewayChange(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable2, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", resourceNameGateway, "id"),
				),
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_disappears(t *testing.T) {
	var rta ec2.RouteTableAssociation
	resourceName := "aws_route_table_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableAssociationConfig_subnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceName, &rta),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceRouteTableAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRouteTableAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route_table_association" {
			continue
		}

		_, err := tfec2.FindRouteTableAssociationByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Route table association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRouteTableAssociationExists(n string, v *ec2.RouteTableAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		association, err := tfec2.FindRouteTableAssociationByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *association

		return nil
	}
}

func testAccRouteTabAssocImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		var target string
		if rs.Primary.Attributes["subnet_id"] != "" {
			target = rs.Primary.Attributes["subnet_id"]
		} else if rs.Primary.Attributes["gateway_id"] != "" {
			target = rs.Primary.Attributes["gateway_id"]
		}
		return fmt.Sprintf("%s/%s", target, rs.Primary.Attributes["route_table_id"]), nil
	}
}

func testAccRouteTableAssociationConfigBaseVPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.1.0/24"

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
`, rName)
}

func testAccVPCRouteTableAssociationConfig_subnet(rName string) string {
	return acctest.ConfigCompose(testAccRouteTableAssociationConfigBaseVPC(rName), fmt.Sprintf(`
resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}
`, rName))
}

func testAccVPCRouteTableAssociationConfig_subnetChange(rName string) string {
	return acctest.ConfigCompose(testAccRouteTableAssociationConfigBaseVPC(rName), fmt.Sprintf(`
resource "aws_route_table" "test2" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test2.id
  subnet_id      = aws_subnet.test.id
}
`, rName))
}

func testAccVPCRouteTableAssociationConfig_gateway(rName string) string {
	return acctest.ConfigCompose(testAccRouteTableAssociationConfigBaseVPC(rName), fmt.Sprintf(`
resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block           = aws_subnet.test.cidr_block
    network_interface_id = aws_network_interface.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  gateway_id     = aws_internet_gateway.test.id
}
`, rName))
}

func testAccVPCRouteTableAssociationConfig_gatewayChange(rName string) string {
	return acctest.ConfigCompose(testAccRouteTableAssociationConfigBaseVPC(rName), fmt.Sprintf(`
resource "aws_route_table" "test2" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block           = aws_subnet.test.cidr_block
    network_interface_id = aws_network_interface.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test2.id
  gateway_id     = aws_internet_gateway.test.id
}
`, rName))
}
