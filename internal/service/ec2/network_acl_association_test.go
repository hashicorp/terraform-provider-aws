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

func TestAccEC2NetworkACLAssociation_basic(t *testing.T) {
	var v ec2.NetworkAclAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_network_acl_association.test"
	naclResourceName := "aws_network_acl.test"
	subnetResourceName := "aws_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLAssociationConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "network_acl_id", naclResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
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

func TestAccEC2NetworkACLAssociation_disappears(t *testing.T) {
	var v ec2.NetworkAclAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_network_acl_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLAssociationConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLAssociationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceNetworkACLAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNetworkACLAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_network_acl_association" {
			continue
		}

		_, err := tfec2.FindNetworkACLAssociationByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Network ACL Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckNetworkACLAssociationExists(n string, v *ec2.NetworkAclAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network ACL Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindNetworkACLAssociationByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccNetworkACLAssociationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id

  cidr_block = "10.1.33.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_association" "test" {
  network_acl_id = aws_network_acl.test.id
  subnet_id      = aws_subnet.test.id
}
`, rName)
}

func testAccNetworkACLAssociationTwoAssociationsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id     = aws_vpc.test.id

  cidr_block = "10.1.33.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id     = aws_vpc.test.id

  cidr_block = "10.1.34.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_association" "test1" {
  network_acl_id = aws_network_acl.test.id
  subnet_id      = aws_subnet.test1.id
}

resource "aws_network_acl_association" "test2" {
  network_acl_id = aws_network_acl.test.id
  subnet_id      = aws_subnet.test2.id
}
`, rName)
}
