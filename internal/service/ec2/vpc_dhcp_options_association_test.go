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

func TestAccVPCDHCPOptionsAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_dhcp_options_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDHCPOptionsAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCDHCPOptionsAssociationExist(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVPCDHCPOptionsAssociationVPCImportIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCDHCPOptionsAssociation_Disappears_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_dhcp_options_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDHCPOptionsAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCDHCPOptionsAssociationExist(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPC(), "aws_vpc.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCDHCPOptionsAssociation_Disappears_dhcp(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_dhcp_options_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDHCPOptionsAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCDHCPOptionsAssociationExist(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCDHCPOptions(), "aws_vpc_dhcp_options.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCDHCPOptionsAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_dhcp_options_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDHCPOptionsAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCDHCPOptionsAssociationExist(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCDHCPOptionsAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCDHCPOptionsAssociation_default(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_dhcp_options_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDHCPOptionsAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDHCPOptionsAssociationConfig_default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCDHCPOptionsAssociationExist(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVPCDHCPOptionsAssociationVPCImportIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVPCDHCPOptionsAssociationVPCImportIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrVPCID], nil
	}
}

func testAccCheckVPCDHCPOptionsAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_dhcp_options_association" {
				continue
			}

			dhcpOptionsID, vpcID, err := tfec2.VPCDHCPOptionsAssociationParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			err = tfec2.FindVPCDHCPOptionsAssociation(ctx, conn, vpcID, dhcpOptionsID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPC DHCP Options Set Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCDHCPOptionsAssociationExist(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC DHCP Options Set Association ID is set")
		}

		dhcpOptionsID, vpcID, err := tfec2.VPCDHCPOptionsAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		return tfec2.FindVPCDHCPOptionsAssociation(ctx, conn, vpcID, dhcpOptionsID)
	}
}

func testAccVPCDHCPOptionsAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_dhcp_options" "test" {
  domain_name          = "service.consul"
  domain_name_servers  = ["127.0.0.1", "10.0.0.2"]
  ntp_servers          = ["127.0.0.1"]
  netbios_name_servers = ["127.0.0.1"]
  netbios_node_type    = 2

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_dhcp_options_association" "test" {
  vpc_id          = aws_vpc.test.id
  dhcp_options_id = aws_vpc_dhcp_options.test.id
}
`, rName)
}

func testAccVPCDHCPOptionsAssociationConfig_default(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_dhcp_options_association" "test" {
  vpc_id          = aws_vpc.test.id
  dhcp_options_id = "default"
}
`, rName)
}
