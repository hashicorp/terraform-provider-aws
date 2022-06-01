package apigatewayv2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
)

func TestAccAPIGatewayV2VPCLink_basic(t *testing.T) {
	var v apigatewayv2.GetVpcLinkOutput
	resourceName := "aws_apigatewayv2_vpc_link.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCLinkConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccVPCLinkConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
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

func TestAccAPIGatewayV2VPCLink_disappears(t *testing.T) {
	var v apigatewayv2.GetVpcLinkOutput
	resourceName := "aws_apigatewayv2_vpc_link.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName, &v),
					testAccCheckVPCLinkDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2VPCLink_tags(t *testing.T) {
	var v apigatewayv2.GetVpcLinkOutput
	resourceName := "aws_apigatewayv2_vpc_link.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCLinkConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckVPCLinkDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_vpc_link" {
			continue
		}

		_, err := conn.GetVpcLink(&apigatewayv2.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 VPC Link %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVPCLinkDisappears(v *apigatewayv2.GetVpcLinkOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		if _, err := conn.DeleteVpcLink(&apigatewayv2.DeleteVpcLinkInput{
			VpcLinkId: v.VpcLinkId,
		}); err != nil {
			return err
		}

		_, err := tfapigatewayv2.WaitVPCLinkDeleted(conn, aws.StringValue(v.VpcLinkId))
		if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error waiting for API Gateway v2 VPC Link (%s) deletion: %s", aws.StringValue(v.VpcLinkId), err)
		}

		return nil
	}
}

func testAccCheckVPCLinkExists(n string, v *apigatewayv2.GetVpcLinkOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 VPC Link ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		resp, err := conn.GetVpcLink(&apigatewayv2.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccVPCLinkConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCLinkConfig_basic(rName string) string {
	return testAccVPCLinkConfig_base(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_vpc_link" "test" {
  name               = %[1]q
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test.*.id
}
`, rName)
}

func testAccVPCLinkConfig_tags(rName string) string {
	return testAccVPCLinkConfig_base(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_vpc_link" "test" {
  name               = %[1]q
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test.*.id

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}
`, rName)
}
