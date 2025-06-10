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

func TestAccVPCEndpointSubnetAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint_subnet_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointSubnetAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointSubnetAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointSubnetAssociationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccVPCEndpointSubnetAssociationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCEndpointSubnetAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint_subnet_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointSubnetAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointSubnetAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointSubnetAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCEndpointSubnetAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCEndpointSubnetAssociation_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName0 := "aws_vpc_endpoint_subnet_association.test.0"
	resourceName1 := "aws_vpc_endpoint_subnet_association.test.1"
	resourceName2 := "aws_vpc_endpoint_subnet_association.test.2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointSubnetAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointSubnetAssociationConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointSubnetAssociationExists(ctx, resourceName0),
					testAccCheckVPCEndpointSubnetAssociationExists(ctx, resourceName1),
					testAccCheckVPCEndpointSubnetAssociationExists(ctx, resourceName2),
				),
			},
		},
	})
}

func testAccCheckVPCEndpointSubnetAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_endpoint_subnet_association" {
				continue
			}

			err := tfec2.FindVPCEndpointSubnetAssociationExists(ctx, conn, rs.Primary.Attributes[names.AttrVPCEndpointID], rs.Primary.Attributes[names.AttrSubnetID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Endpoint Subnet Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCEndpointSubnetAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		err := tfec2.FindVPCEndpointSubnetAssociationExists(ctx, conn, rs.Primary.Attributes[names.AttrVPCEndpointID], rs.Primary.Attributes[names.AttrSubnetID])

		if err != nil {
			return err
		}

		return err
	}
}

func testAccVPCEndpointSubnetAssociationConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 3), fmt.Sprintf(`
data "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = "default"
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  vpc_endpoint_type   = "Interface"
  service_name        = "com.amazonaws.${data.aws_region.current.name}.ec2"
  security_group_ids  = [data.aws_security_group.test.id]
  private_dns_enabled = false

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointSubnetAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointSubnetAssociationConfig_base(rName), `
resource "aws_vpc_endpoint_subnet_association" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.test.id
  subnet_id       = aws_subnet.test[0].id
}
`)
}

func testAccVPCEndpointSubnetAssociationConfig_multiple(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointSubnetAssociationConfig_base(rName), `
resource "aws_vpc_endpoint_subnet_association" "test" {
  count = 3

  vpc_endpoint_id = aws_vpc_endpoint.test.id
  subnet_id       = aws_subnet.test[count.index].id
}
`)
}

func testAccVPCEndpointSubnetAssociationImportStateIdFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		id := fmt.Sprintf("%s/%s", rs.Primary.Attributes[names.AttrVPCEndpointID], rs.Primary.Attributes[names.AttrSubnetID])
		return id, nil
	}
}
