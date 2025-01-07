// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
)

func TestAccVPCLatticeResourceGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var resourcegateway vpclattice.GetResourceGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn", "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
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

func TestAccVPCLatticeResourceGateway_addressTypeDualstack(t *testing.T) {
	ctx := acctest.Context(t)

	var resourcegateway vpclattice.GetResourceGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_gateway.test"
	addressType := "DUALSTACK"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_addressType(rName, addressType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", addressType),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn", "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
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

func TestAccVPCLatticeResourceGateway_addressTypeIPv6(t *testing.T) {
	ctx := acctest.Context(t)

	var resourcegateway vpclattice.GetResourceGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_gateway.test"
	addressType := "IPV6"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_addressType(rName, addressType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", addressType),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn", "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
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

func TestAccVPCLatticeResourceGateway_update(t *testing.T) {
	ctx := acctest.Context(t)

	var resourcegateway vpclattice.GetResourceGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_gateway.test"
	securityGroup1 := "aws_security_group.test"
	securityGroup2 := "aws_security_group.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup1, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn", "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceGatewayConfig_update1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup1, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup2, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn", "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
				),
			},
			{
				Config: testAccResourceGatewayConfig_update2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup2, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn", "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
				),
			},
		},
	})
}

func TestAccVPCLatticeResourceGateway_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourcegateway vpclattice.GetResourceGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceResourceGateway = newResourceResourceGateway
					// acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceResourceGateway, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourceGatewayDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_resource_gateway" {
				continue
			}

			input := &vpclattice.GetResourceGatewayInput{
				ResourceGatewayIdentifier: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetResourceGateway(ctx, input)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameResourceGateway, rs.Primary.ID, err)
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameResourceGateway, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResourceGatewayExists(ctx context.Context, name string, resourcegateway *vpclattice.GetResourceGatewayOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameResourceGateway, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameResourceGateway, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)
		resp, err := conn.GetResourceGateway(ctx, &vpclattice.GetResourceGatewayInput{
			ResourceGatewayIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameResourceGateway, rs.Primary.ID, err)
		}

		*resourcegateway = *resp

		return nil
	}
}

func testAccCheckResourceGatewayNotRecreated(before, after *vpclattice.GetResourceGatewayOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return create.Error(names.VPCLattice, create.ErrActionCheckingNotRecreated, tfvpclattice.ResNameResourceGateway, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccResourceGatewayConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"
  ipv6_cidr_block = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 0)

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name = %[1]q
  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccResourceGatewayConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_resource_gateway" "test" {
  name             = %[1]q
  vpc_id = aws_vpc.test.id
  security_group_ids         = [aws_security_group.test.id]
  subnet_ids = [aws_subnet.test.id]
  ip_address_type = "IPV4"
}
`, rName))
}

func testAccResourceGatewayConfig_addressType(rName, addressType string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_resource_gateway" "test" {
  name             = %[1]q
  vpc_id = aws_vpc.test.id
  security_group_ids         = [aws_security_group.test.id]
  subnet_ids = [aws_subnet.test.id]
  ip_address_type = %[2]q
}
`, rName, addressType))
}

func testAccResourceGatewayConfig_multipleSubnets(rName string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_resource_gateway" "test" {
  name             = %[1]q
  vpc_id = aws_vpc.test.id
  security_group_ids         = [aws_security_group.test.id]
  subnet_ids = [aws_subnet.test.id, aws_subnet.test2.id]
  ip_address_type = "IPV4"
}
`, rName))
}

func testAccResourceGatewayConfig_update1(rName string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_vpclattice_resource_gateway" "test" {
  name             = %[1]q
  vpc_id = aws_vpc.test.id
  security_group_ids         = [aws_security_group.test.id, aws_security_group.test2.id]
  subnet_ids = [aws_subnet.test.id]
  ip_address_type = "IPV4"
}
`, rName))
}

func testAccResourceGatewayConfig_update2(rName string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_vpclattice_resource_gateway" "test" {
  name             = %[1]q
  vpc_id = aws_vpc.test.id
  security_group_ids         = [aws_security_group.test2.id]
  subnet_ids = [aws_subnet.test.id]
  ip_address_type = "IPV4"
}
`, rName))
}
