// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeServiceNetworkResourceAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var servicenetworkresourceassociation vpclattice.GetServiceNetworkResourceAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_resource_association.test"
	resourceServiceNetworkName := "aws_vpclattice_service_network.test"
	resourceConfigurationName := "aws_vpclattice_resource_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkResourceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkResourceAssociationExists(ctx, resourceName, &servicenetworkresourceassociation),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var servicenetworkresourceassociation vpclattice.GetServiceNetworkResourceAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_resource_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkResourceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkResourceAssociationExists(ctx, resourceName, &servicenetworkresourceassociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceServiceNetworkResourceAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceNetworkResourceAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_service_network_resource_association" {
				continue
			}

			_, err := tfvpclattice.FindServiceNetworkResourceAssociationByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameServiceNetworkResourceAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameServiceNetworkResourceAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckServiceNetworkResourceAssociationExists(ctx context.Context, name string, servicenetworkresourceassociation *vpclattice.GetServiceNetworkResourceAssociationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameServiceNetworkResourceAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameServiceNetworkResourceAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		resp, err := tfvpclattice.FindServiceNetworkResourceAssociationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameServiceNetworkResourceAssociation, rs.Primary.ID, err)
		}

		*servicenetworkresourceassociation = *resp

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
