// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeServiceNetworkResourceAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var servicenetworkresourceassociation vpclattice.GetServiceNetworkResourceAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_resource_association.test"
	resourceServiceNetworkName := "aws_vpclattice_service_network.test"
	resourceConfigurationName := "aws_vpclattice_resource_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkResourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkResourceAssociationExists(ctx, t, resourceName, &servicenetworkresourceassociation),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "resource_configuration_identifier", resourceConfigurationName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "service_network_identifier", resourceServiceNetworkName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "dns_entry.0.domain_name"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_entry.0.hosted_zone_id"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`servicenetworkresourceassociation/+.`)),
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

func TestAccVPCLatticeServiceNetworkResourceAssociation_privateDNS(t *testing.T) {
	ctx := acctest.Context(t)
	var servicenetworkresourceassociation vpclattice.GetServiceNetworkResourceAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_resource_association.test"
	resourceServiceNetworkName := "aws_vpclattice_service_network.test"
	resourceConfigurationName := "aws_vpclattice_resource_configuration.test"
	domainName := fmt.Sprintf("%s.example.com", rName)
	customDomainName := fmt.Sprintf("test.%s.example.com", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkResourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkResourceAssociationConfig_privateDNS(rName, domainName, customDomainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkResourceAssociationExists(ctx, t, resourceName, &servicenetworkresourceassociation),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "resource_configuration_identifier", resourceConfigurationName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "service_network_identifier", resourceServiceNetworkName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "dns_entry.0.domain_name"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_entry.0.hosted_zone_id"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`servicenetworkresourceassociation/+.`)),
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

func TestAccVPCLatticeServiceNetworkResourceAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var servicenetworkresourceassociation vpclattice.GetServiceNetworkResourceAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_resource_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkResourceAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkResourceAssociationExists(ctx, t, resourceName, &servicenetworkresourceassociation),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfvpclattice.ResourceServiceNetworkResourceAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceNetworkResourceAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_service_network_resource_association" {
				continue
			}

			_, err := tfvpclattice.FindServiceNetworkResourceAssociationByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Lattice Service Network Resource Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServiceNetworkResourceAssociationExists(ctx context.Context, t *testing.T, n string, v *vpclattice.GetServiceNetworkResourceAssociationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).VPCLatticeClient(ctx)

		output, err := tfvpclattice.FindServiceNetworkResourceAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccServiceNetworkResourceAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccResourceConfigurationConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_service_network_resource_association" "test" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.test.id
  service_network_identifier        = aws_vpclattice_service_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServiceNetworkResourceAssociationConfig_privateDNS(rName, domainName, customDomainName string) string {
	return acctest.ConfigCompose(testAccResourceConfigurationConfig_domainVerification(rName, domainName, customDomainName), fmt.Sprintf(`
resource "aws_vpclattice_service_network_resource_association" "test" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.test.id
  service_network_identifier        = aws_vpclattice_service_network.test.id

  private_dns_enabled = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
