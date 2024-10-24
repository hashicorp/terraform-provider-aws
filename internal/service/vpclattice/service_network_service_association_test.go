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

func TestAccVPCLatticeServiceNetworkServiceAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var servicenetworkasc vpclattice.GetServiceNetworkServiceAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_service_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkServiceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkServiceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkServiceAssociationExists(ctx, resourceName, &servicenetworkasc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile("servicenetworkserviceassociation/.+$")),
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

func TestAccVPCLatticeServiceNetworkServiceAssociation_arn(t *testing.T) {
	ctx := acctest.Context(t)

	var servicenetworkasc vpclattice.GetServiceNetworkServiceAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_service_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkServiceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkServiceAssociationConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkServiceAssociationExists(ctx, resourceName, &servicenetworkasc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile("servicenetworkserviceassociation/.+$")),
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

func TestAccVPCLatticeServiceNetworkServiceAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var servicenetworkasc vpclattice.GetServiceNetworkServiceAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_service_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkServiceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkServiceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkServiceAssociationExists(ctx, resourceName, &servicenetworkasc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceServiceNetworkServiceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCLatticeServiceNetworkServiceAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var servicenetworkasc1, servicenetworkasc2, service3 vpclattice.GetServiceNetworkServiceAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_service_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkServiceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkServiceAssociationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkServiceAssociationExists(ctx, resourceName, &servicenetworkasc1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceNetworkServiceAssociationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkServiceAssociationExists(ctx, resourceName, &servicenetworkasc2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccServiceNetworkServiceAssociationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkServiceAssociationExists(ctx, resourceName, &service3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckServiceNetworkServiceAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_service_network_service_association" {
				continue
			}

			_, err := tfvpclattice.FindServiceNetworkServiceAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Lattice Service Network Service Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServiceNetworkServiceAssociationExists(ctx context.Context, name string, service *vpclattice.GetServiceNetworkServiceAssociationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameService, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameService, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)
		resp, err := tfvpclattice.FindServiceNetworkServiceAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*service = *resp

		return nil
	}
}

func testAccServiceNetworkServiceAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}
`, rName)
}

func testAccServiceNetworkServiceAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccServiceNetworkServiceAssociationConfig_base(rName), `
resource "aws_vpclattice_service_network_service_association" "test" {
  service_identifier         = aws_vpclattice_service.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id
}
`)
}

func testAccServiceNetworkServiceAssociationConfig_arn(rName string) string {
	return acctest.ConfigCompose(testAccServiceNetworkServiceAssociationConfig_base(rName), `
resource "aws_vpclattice_service_network_service_association" "test" {
  service_identifier         = aws_vpclattice_service.test.arn
  service_network_identifier = aws_vpclattice_service_network.test.arn
}
`)
}

func testAccServiceNetworkServiceAssociationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccServiceNetworkServiceAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_service_network_service_association" "test" {
  service_identifier         = aws_vpclattice_service.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccServiceNetworkServiceAssociationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccServiceNetworkServiceAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_service_network_service_association" "test" {
  service_identifier         = aws_vpclattice_service.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
