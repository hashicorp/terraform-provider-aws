package apigateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
)

func TestAccAPIGatewayVPCLink_basic(t *testing.T) {
	rName := sdkacctest.RandString(5)
	resourceName := "aws_api_gateway_vpc_link.test"
	vpcLinkName := fmt.Sprintf("tf-apigateway-%s", rName)
	vpcLinkNameUpdated := fmt.Sprintf("tf-apigateway-update-%s", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayVpcLinkConfig(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", vpcLinkName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIGatewayVpcLinkConfig_Update(rName, "test update"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/vpclinks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", vpcLinkNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "test update"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccAPIGatewayVPCLink_tags(t *testing.T) {
	rName := sdkacctest.RandString(5)
	resourceName := "aws_api_gateway_vpc_link.test"
	vpcLinkName := fmt.Sprintf("tf-apigateway-%s", rName)
	description := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayVpcLinkConfigTags1(rName, description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", vpcLinkName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", "1"),
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
				Config: testAccAPIGatewayVpcLinkConfigTags2(rName, description, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", vpcLinkName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAPIGatewayVpcLinkConfigTags1(rName, description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", vpcLinkName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "target_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayVPCLink_disappears(t *testing.T) {
	rName := sdkacctest.RandString(5)
	resourceName := "aws_api_gateway_vpc_link.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayVpcLinkConfig(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCLinkExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceVPCLink(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCLinkDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_vpc_link" {
			continue
		}

		input := &apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetVpcLink(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected VPC Link to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVPCLinkExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		input := &apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetVpcLink(input)
		return err
	}
}

func testAccAPIGatewayVpcLinkConfig_basis(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test_a" {
  name               = "tf-lb-%s"
  internal           = true
  load_balancer_type = "network"
  subnets            = [aws_subnet.test.id]
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "tf-acc-api-gateway-vpc-link-%s"
  }
}

data "aws_availability_zones" "test" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.10.0.0/21"
  availability_zone = data.aws_availability_zones.test.names[0]

  tags = {
    Name = "tf-acc-api-gateway-vpc-link"
  }
}
`, rName, rName)
}

func testAccAPIGatewayVpcLinkConfig(rName, description string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name        = "tf-apigateway-%s"
  description = %q
  target_arns = [aws_lb.test_a.arn]
}
`, rName, description)
}

func testAccAPIGatewayVpcLinkConfigTags1(rName, description, tagKey1, tagValue1 string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name        = "tf-apigateway-%s"
  description = %q
  target_arns = [aws_lb.test_a.arn]

  tags = {
    %q = %q
  }
}
`, rName, description, tagKey1, tagValue1)
}

func testAccAPIGatewayVpcLinkConfigTags2(rName, description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name        = "tf-apigateway-%s"
  description = %q
  target_arns = [aws_lb.test_a.arn]

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, description, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAPIGatewayVpcLinkConfig_Update(rName, description string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name        = "tf-apigateway-update-%s"
  description = %q
  target_arns = [aws_lb.test_a.arn]
}
`, rName, description)
}
