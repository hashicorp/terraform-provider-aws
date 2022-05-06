package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccVPCEndpointPolicy_basic(t *testing.T) {
	var endpoint ec2.VpcEndpoint

	resourceName := "aws_vpc_endpoint_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointPolicyBasicConfig(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcEndpointPolicyBasicConfig(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
				),
			},
		},
	})
}

func TestAccVPCEndpointPolicy_disappears(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointPolicyBasicConfig(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPCEndpointPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCEndpointPolicy_disappears_endpoint(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointPolicyBasicConfig(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPCEndpoint(), "aws_vpc_endpoint.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

const policy1 = `
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadOnly",
      "Principal": "*",
      "Action": [
        "dynamodb:DescribeTable",
        "dynamodb:ListTables"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
`

const policy2 = `
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}
`

func testAccVpcEndpointPolicyBasicConfig(rName, policy string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service = "dynamodb"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_name = data.aws_vpc_endpoint_service.test.service_name
  vpc_id       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_policy" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.test.id
  policy          = <<POLICY
%[2]s
POLICY
}
`, rName, policy)
}
