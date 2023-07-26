// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccClientVPNNetworkAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assoc ec2.TargetNetwork
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_network_association.test"
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNNetworkAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNNetworkAssociationConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNNetworkAssociationExists(ctx, resourceName, &assoc),
					resource.TestMatchResourceAttr(resourceName, "association_id", regexp.MustCompile("^cvpn-assoc-[a-z0-9]+$")),
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "client_vpn_endpoint_id", endpointResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccClientVPNNetworkAssociationImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccClientVPNNetworkAssociation_multipleSubnets(t *testing.T) {
	ctx := acctest.Context(t)
	var assoc ec2.TargetNetwork
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceNames := []string{"aws_ec2_client_vpn_network_association.test.0", "aws_ec2_client_vpn_network_association.test.1"}
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceNames := []string{"aws_subnet.test.0", "aws_subnet.test.1"}
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNNetworkAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNNetworkAssociationConfig_multipleSubnets(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNNetworkAssociationExists(ctx, resourceNames[0], &assoc),
					resource.TestMatchResourceAttr(resourceNames[0], "association_id", regexp.MustCompile("^cvpn-assoc-[a-z0-9]+$")),
					resource.TestMatchResourceAttr(resourceNames[1], "association_id", regexp.MustCompile("^cvpn-assoc-[a-z0-9]+$")),
					resource.TestCheckResourceAttrPair(resourceNames[0], "id", resourceNames[0], "association_id"),
					resource.TestCheckResourceAttrPair(resourceNames[0], "client_vpn_endpoint_id", endpointResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceNames[0], "subnet_id", subnetResourceNames[0], "id"),
					resource.TestCheckResourceAttrPair(resourceNames[1], "subnet_id", subnetResourceNames[1], "id"),
					resource.TestCheckResourceAttrPair(resourceNames[0], "vpc_id", vpcResourceName, "id"),
				),
			},
		},
	})
}

func testAccClientVPNNetworkAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var assoc ec2.TargetNetwork
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_network_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNNetworkAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNNetworkAssociationConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNNetworkAssociationExists(ctx, resourceName, &assoc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceClientVPNNetworkAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClientVPNNetworkAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_client_vpn_network_association" {
				continue
			}

			_, err := tfec2.FindClientVPNNetworkAssociationByIDs(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["client_vpn_endpoint_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Client VPN Network Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClientVPNNetworkAssociationExists(ctx context.Context, name string, v *ec2.TargetNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Client VPN Network Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		output, err := tfec2.FindClientVPNNetworkAssociationByIDs(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["client_vpn_endpoint_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClientVPNNetworkAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["client_vpn_endpoint_id"], rs.Primary.ID), nil
	}
}

func testAccClientVPNNetworkAssociationConfig_base(t *testing.T, rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), testAccClientVPNEndpointConfig_basic(t, rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccClientVPNNetworkAssociationConfig_basic(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNNetworkAssociationConfig_base(t, rName), `
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test[0].id
}
`)
}

func testAccClientVPNNetworkAssociationConfig_multipleSubnets(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNNetworkAssociationConfig_base(t, rName), `
resource "aws_ec2_client_vpn_network_association" "test" {
  count = 2

  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test[count.index].id
}
`)
}
