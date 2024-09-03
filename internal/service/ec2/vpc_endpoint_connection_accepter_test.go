// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCEndpointConnectionAccepter_crossAccount(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint_connection_accepter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckVPCEndpointConnectionAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConnectionAccepterConfig_crossAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_state", "available"),
				),
			},
			{
				Config:            testAccVPCEndpointConnectionAccepterConfig_crossAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCEndpointConnectionAccepterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_endpoint_connection_accepter" {
				continue
			}

			_, err := tfec2.FindVPCEndpointConnectionByServiceIDAndVPCEndpointID(ctx, conn, rs.Primary.Attributes["vpc_endpoint_service_id"], rs.Primary.Attributes[names.AttrVPCEndpointID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Endpoint Connection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVPCEndpointConnectionAccepterConfig_crossAccount(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
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

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test3" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

resource "aws_vpc" "test_alternate" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "alternate_available" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test_alternate1" {
  provider = "awsalternate"

  vpc_id            = aws_vpc.test_alternate.id
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.alternate_available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_alternate2" {
  provider = "awsalternate"

  vpc_id            = aws_vpc.test_alternate.id
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.alternate_available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_alternate3" {
  provider = "awsalternate"

  vpc_id            = aws_vpc.test_alternate.id
  cidr_block        = "10.1.3.0/24"
  availability_zone = data.aws_availability_zones.alternate_available.names[2]

  tags = {
    Name = %[1]q
  }
}

data "aws_subnets" "alternate_intersect" {
  provider = "awsalternate"

  filter {
    name   = "availabilityZone"
    values = aws_vpc_endpoint_service.test.availability_zones
  }

  filter {
    name   = "vpc-id"
    values = [aws_vpc.test_alternate.id]
  }
}

resource "aws_lb" "test" {
  name = %[1]q

  subnets = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
    aws_subnet.test3.id,
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = true

  network_load_balancer_arns = [
    aws_lb.test.id,
  ]

  allowed_principals = [
    "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.alternate.account_id}:root",
  ]
}

resource "aws_security_group" "test" {
  provider = "awsalternate"

  vpc_id = aws_vpc.test_alternate.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  provider = "awsalternate"

  vpc_id              = aws_vpc.test_alternate.id
  service_name        = aws_vpc_endpoint_service.test.service_name
  subnet_ids          = data.aws_subnets.alternate_intersect.ids
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false

  security_group_ids = [
    aws_security_group.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_connection_accepter" "test" {
  vpc_endpoint_service_id = aws_vpc_endpoint_service.test.id
  vpc_endpoint_id         = aws_vpc_endpoint.test.id
}
`, rName))
}
