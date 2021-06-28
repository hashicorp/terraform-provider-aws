package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSVpcEndpointServiceAllowedPrincipal_basic(t *testing.T) {
	resourceName := "aws_vpc_endpoint_service_allowed_principal.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceAllowedPrincipalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceAllowedPrincipalConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceAllowedPrincipalExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckVpcEndpointServiceAllowedPrincipalDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint_service_allowed_principal" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeVpcEndpointServicePermissions(&ec2.DescribeVpcEndpointServicePermissionsInput{
			ServiceId: aws.String(rs.Primary.Attributes["vpc_endpoint_service_id"]),
		})
		if err != nil {
			// Verify the error is what we want
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if ec2err.Code() != "InvalidVpcEndpointServiceId.NotFound" {
				return err
			}
			return nil
		}

		if len(resp.AllowedPrincipals) > 0 {
			return fmt.Errorf(
				"VCP Endpoint Service %s has allowed principals", rs.Primary.Attributes["vpc_endpoint_service_id"])
		}
	}

	return nil
}

func testAccCheckVpcEndpointServiceAllowedPrincipalExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Endpoint Service ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeVpcEndpointServicePermissions(&ec2.DescribeVpcEndpointServicePermissionsInput{
			ServiceId: aws.String(rs.Primary.Attributes["vpc_endpoint_service_id"]),
		})
		if err != nil {
			return err
		}

		for _, principal := range resp.AllowedPrincipals {
			if aws.StringValue(principal.Principal) == rs.Primary.Attributes["principal_arn"] {
				return nil
			}
		}

		return fmt.Errorf("VPC Endpoint Service allowed principal not found")
	}
}

func testAccVpcEndpointServiceAllowedPrincipalConfig(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(
		`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name = %[1]q

  subnets = [
    aws_subnet.test[0].id,
    aws_subnet.test[1].id,
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
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

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = false

  network_load_balancer_arns = [
    aws_lb.test.arn,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service_allowed_principal" "test" {
  vpc_endpoint_service_id = aws_vpc_endpoint_service.test.id

  principal_arn = data.aws_iam_session_context.current.issuer_arn
}
`, rName))
}
