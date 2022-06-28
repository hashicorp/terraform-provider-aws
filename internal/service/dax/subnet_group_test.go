package dax_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDAXSubnetGroup_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dax_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dax.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists("aws_dax_subnet_group.test"),
					resource.TestCheckResourceAttr("aws_dax_subnet_group.test", "subnet_ids.#", "2"),
					resource.TestCheckResourceAttrSet("aws_dax_subnet_group.test", "vpc_id"),
				),
			},
			{
				Config: testAccSubnetGroupConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists("aws_dax_subnet_group.test"),
					resource.TestCheckResourceAttr("aws_dax_subnet_group.test", "description", "update"),
					resource.TestCheckResourceAttr("aws_dax_subnet_group.test", "subnet_ids.#", "3"),
					resource.TestCheckResourceAttrSet("aws_dax_subnet_group.test", "vpc_id"),
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

func testAccCheckSubnetGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DAXConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dax_subnet_group" {
			continue
		}

		_, err := conn.DescribeSubnetGroups(&dax.DescribeSubnetGroupsInput{
			SubnetGroupNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, dax.ErrCodeSubnetGroupNotFoundFault) {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckSubnetGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DAXConn

		_, err := conn.DescribeSubnetGroups(&dax.DescribeSubnetGroupsInput{
			SubnetGroupNames: []*string{aws.String(rs.Primary.ID)},
		})

		return err
	}
}

func testAccSubnetGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  cidr_block = "10.0.1.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_subnet" "test2" {
  cidr_block = "10.0.2.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_dax_subnet_group" "test" {
  name = "%s"

  subnet_ids = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
  ]
}
`, rName)
}

func testAccSubnetGroupConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  cidr_block = "10.0.1.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_subnet" "test2" {
  cidr_block = "10.0.2.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_subnet" "test3" {
  cidr_block = "10.0.3.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_dax_subnet_group" "test" {
  name        = "%s"
  description = "update"

  subnet_ids = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
    aws_subnet.test3.id,
  ]
}
`, rName)
}
