package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsEc2LocalGatewayRouteTableVpcAssociation_basic(t *testing.T) {
	// Hide Outposts testing behind consistent environment variable
	outpostArn := os.Getenv("AWS_OUTPOST_ARN")
	if outpostArn == "" {
		t.Skip(
			"Environment variable AWS_OUTPOST_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"a deployed Outpost to enable this test.")
	}

	rName := acctest.RandomWithPrefix("tf-acc-test")
	localGatewayRouteTableDataSourceName := "data.aws_ec2_local_gateway_route_table.test"
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfig(rName, outpostArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName),
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

func TestAccAwsEc2LocalGatewayRouteTableVpcAssociation_disappears(t *testing.T) {
	// Hide Outposts testing behind consistent environment variable
	outpostArn := os.Getenv("AWS_OUTPOST_ARN")
	if outpostArn == "" {
		t.Skip(
			"Environment variable AWS_OUTPOST_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"a deployed Outpost to enable this test.")
	}

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfig(rName, outpostArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2LocalGatewayRouteTableVpcAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsEc2LocalGatewayRouteTableVpcAssociation_Tags(t *testing.T) {
	// Hide Outposts testing behind consistent environment variable
	outpostArn := os.Getenv("AWS_OUTPOST_ARN")
	if outpostArn == "" {
		t.Skip(
			"Environment variable AWS_OUTPOST_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"a deployed Outpost to enable this test.")
	}

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigTags1(rName, outpostArn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName),
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
				Config: testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigTags2(rName, outpostArn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigTags1(rName, outpostArn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Fleet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		association, err := getEc2LocalGatewayRouteTableVpcAssociation(conn, rs.Primary.ID)

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

func testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_local_gateway_route_table_vpc_association" {
			continue
		}

		association, err := getEc2LocalGatewayRouteTableVpcAssociation(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if association != nil && aws.StringValue(association.State) != ec2.RouteTableAssociationStateCodeDisassociated {
			return fmt.Errorf("EC2 Local Gateway Route Table VPC Association (%s) still exists in state: %s", rs.Primary.ID, aws.StringValue(association.State))
		}
	}

	return nil
}

func testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigBase(rName, outpostArn string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = %[2]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, outpostArn)
}

func testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfig(rName, outpostArn string) string {
	return composeConfig(
		testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigBase(rName, outpostArn),
		`
resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id
}
`)
}

func testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigTags1(rName, outpostArn, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigBase(rName, outpostArn),
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

func testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigTags2(rName, outpostArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigBase(rName, outpostArn),
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
