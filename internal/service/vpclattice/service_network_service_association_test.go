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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkServiceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkServiceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkServiceAssociationExists(ctx, resourceName, &servicenetworkasc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile("servicenetworkserviceassociation/.+$")),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkServiceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkServiceAssociationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkServiceAssociationExists(ctx, resourceName, &servicenetworkasc1),
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
				Config: testAccServiceNetworkServiceAssociationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkServiceAssociationExists(ctx, resourceName, &servicenetworkasc2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccServiceNetworkServiceAssociationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkServiceAssociationExists(ctx, resourceName, &service3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

			_, err := conn.GetServiceNetworkServiceAssociation(ctx, &vpclattice.GetServiceNetworkServiceAssociationInput{
				ServiceNetworkServiceAssociationIdentifier: aws.String(rs.Primary.ID),
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
		resp, err := conn.GetServiceNetworkServiceAssociation(ctx, &vpclattice.GetServiceNetworkServiceAssociationInput{
			ServiceNetworkServiceAssociationIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameService, rs.Primary.ID, err)
		}

		*service = *resp

		return nil
	}
}

func testAccServiceNetworkServiceAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network_service_association" "test" {
  service_identifier         = aws_vpclattice_service.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id
}
`, rName)
}

func testAccServiceNetworkServiceAssociationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network_service_association" "test" {
  service_identifier         = aws_vpclattice_service.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccServiceNetworkServiceAssociationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_vpclattice_service_network_service_association" "test" {
  service_identifier         = aws_vpclattice_service.test.id
  service_network_identifier = aws_vpclattice_service_network.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
