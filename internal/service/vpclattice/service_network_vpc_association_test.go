// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &servicenetworkvpcasc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile("servicenetworkvpcassociation/.+$")),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkVPCAssociationConfig_full(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &servicenetworkvpcasc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile("servicenetworkvpcassociation/.+$")),
					resource.TestCheckResourceAttrSet(resourceName, "service_network_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_identifier"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkVPCAssociationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &servicenetworkvpcasc1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceNetworkVPCAssociationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &servicenetworkvpcasc2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccServiceNetworkVPCAssociationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkVPCAssociationExists(ctx, resourceName, &service3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

			_, err := conn.GetServiceNetworkVpcAssociation(ctx, &vpclattice.GetServiceNetworkVpcAssociationInput{
				ServiceNetworkVpcAssociationIdentifier: aws.String(rs.Primary.ID),
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameService, rs.Primary.ID, errors.New("not destroyed"))
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
		resp, err := conn.GetServiceNetworkVpcAssociation(ctx, &vpclattice.GetServiceNetworkVpcAssociationInput{
			ServiceNetworkVpcAssociationIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameService, rs.Primary.ID, err)
		}

		*service = *resp

		return nil
	}
}

func testAccServiceNetworkVPCAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network_vpc_association" "test" {
  vpc_identifier             = aws_vpc.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id
}
`, rName)
}

func testAccServiceNetworkVPCAssociationConfig_full(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network_vpc_association" "test" {
  vpc_identifier             = aws_vpc.test.id
  security_group_ids         = [aws_security_group.test.id]
  service_network_identifier = aws_vpclattice_service_network.test.id
}
`, rName)
}

func testAccServiceNetworkVPCAssociationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network_vpc_association" "test" {
  vpc_identifier             = aws_vpc.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccServiceNetworkVPCAssociationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network_vpc_association" "test" {
  vpc_identifier             = aws_vpc.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
