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

func TestAccVPCLatticeServiceNetworkVPCAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var servicenetworkvpcasc vpclattice.GetServiceNetworkVpcAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &servicenetworkvpcasc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile("servicenetworkvpcassociation/.+$")),
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

func TestAccVPCLatticeServiceNetworkVPCAssociation_arn(t *testing.T) {
	ctx := acctest.Context(t)

	var servicenetworkvpcasc vpclattice.GetServiceNetworkVpcAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkVPCAssociationConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &servicenetworkvpcasc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile("servicenetworkvpcassociation/.+$")),
					resource.TestCheckResourceAttrSet(resourceName, "service_network_identifier"),
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

func TestAccVPCLatticeServiceNetworkVPCAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var servicenetworkvpcasc vpclattice.GetServiceNetworkVpcAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &servicenetworkvpcasc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceServiceNetworkVPCAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCLatticeServiceNetworkVPCAssociation_full(t *testing.T) {
	ctx := acctest.Context(t)

	var servicenetworkvpcasc vpclattice.GetServiceNetworkVpcAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkVPCAssociationConfig_full(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &servicenetworkvpcasc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile("servicenetworkvpcassociation/.+$")),
					resource.TestCheckResourceAttrSet(resourceName, "service_network_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_identifier"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
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

func TestAccVPCLatticeServiceNetworkVPCAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var servicenetworkvpcasc1, servicenetworkvpcasc2, service3 vpclattice.GetServiceNetworkVpcAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkVPCAssociationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &servicenetworkvpcasc1),
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
				Config: testAccServiceNetworkVPCAssociationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &servicenetworkvpcasc2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccServiceNetworkVPCAssociationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &service3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckServiceNetworkVPCAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_service_network_vpc_association" {
				continue
			}

			_, err := tfvpclattice.FindServiceNetworkVPCAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Lattice Service Network VPC Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServiceNetworkVPCAssociationExists(ctx context.Context, name string, service *vpclattice.GetServiceNetworkVpcAssociationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameService, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameService, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)
		resp, err := tfvpclattice.FindServiceNetworkVPCAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*service = *resp

		return nil
	}
}

func testAccServiceNetworkVPCAssociationConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 0), fmt.Sprintf(`
resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}
`, rName))
}

func testAccServiceNetworkVPCAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccServiceNetworkVPCAssociationConfig_base(rName), `
resource "aws_vpclattice_service_network_vpc_association" "test" {
  vpc_identifier             = aws_vpc.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id
}
`)
}

func testAccServiceNetworkVPCAssociationConfig_arn(rName string) string {
	return acctest.ConfigCompose(testAccServiceNetworkVPCAssociationConfig_base(rName), `
resource "aws_vpclattice_service_network_vpc_association" "test" {
  vpc_identifier             = aws_vpc.test.id
  service_network_identifier = aws_vpclattice_service_network.test.arn
}
`)
}

func testAccServiceNetworkVPCAssociationConfig_full(rName string) string {
	return acctest.ConfigCompose(testAccServiceNetworkVPCAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_service_network_vpc_association" "test" {
  vpc_identifier             = aws_vpc.test.id
  security_group_ids         = [aws_security_group.test.id]
  service_network_identifier = aws_vpclattice_service_network.test.id
}
`, rName))
}

func testAccServiceNetworkVPCAssociationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccServiceNetworkVPCAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_service_network_vpc_association" "test" {
  vpc_identifier             = aws_vpc.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccServiceNetworkVPCAssociationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccServiceNetworkVPCAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_service_network_vpc_association" "test" {
  vpc_identifier             = aws_vpc.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
