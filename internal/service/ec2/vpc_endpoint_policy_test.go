// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCEndpointPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint

	resourceName := "aws_vpc_endpoint_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointPolicyConfig_basic(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointPolicyConfig_basic(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
			},
		},
	})
}

func TestAccVPCEndpointPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointPolicyConfig_basic(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCEndpointPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCEndpointPolicy_disappears_endpoint(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointPolicyConfig_basic(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCEndpoint(), "aws_vpc_endpoint.test"),
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

func testAccVPCEndpointPolicyConfig_basic(rName, policy string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service      = "dynamodb"
  service_type = "Gateway"
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
