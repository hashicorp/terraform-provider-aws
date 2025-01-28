// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "IPV4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_addressType(rName, addressType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, addressType),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_addressType(rName, addressType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, addressType),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
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

func TestAccVPCLatticeResourceGateway_multipleSubnets(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcegateway vpclattice.GetResourceGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_gateway.test"
	subnetResourceName1 := "aws_subnet.test"
	subnetResourceName2 := "aws_subnet.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_multipleSubnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "IPV4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", subnetResourceName1, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", subnetResourceName2, names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_update1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "IPV4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup1, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceGatewayConfig_update2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "IPV4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup1, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup2, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "IPV4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup1, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourcegateway/rgw-.+`)),
				),
			},
		},
	})
}

func TestAccVPCLatticeResourceGateway_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcegateway vpclattice.GetResourceGatewayOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGatewayConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGatewayExists(ctx, resourceName, &resourcegateway),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceResourceGateway, resourceName),
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

			_, err := tfvpclattice.FindResourceGatewayByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Lattice Resource Gateway %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourceGatewayExists(ctx context.Context, n string, v *vpclattice.GetResourceGatewayOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		output, err := tfvpclattice.FindResourceGatewayByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccResourceGatewayConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 0)

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccResourceGatewayConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}
`, rName))
}

func testAccResourceGatewayConfig_addressType(rName, addressType string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_resource_gateway" "test" {
  name               = %[1]q
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = [aws_subnet.test.id]
  ip_address_type    = %[2]q
}
`, rName, addressType))
}

func testAccResourceGatewayConfig_multipleSubnets(rName string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_subnet" "test2" {
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_resource_gateway" "test" {
  name               = %[1]q
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  ip_address_type    = "IPV4"
}
`, rName))
}

func testAccResourceGatewayConfig_update1(rName string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_vpclattice_resource_gateway" "test" {
  name               = %[1]q
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = [aws_subnet.test.id]
  ip_address_type    = "IPV4"
}
`, rName))
}

func testAccResourceGatewayConfig_update2(rName string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_vpclattice_resource_gateway" "test" {
  name               = %[1]q
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
  subnet_ids         = [aws_subnet.test.id]
  ip_address_type    = "IPV4"
}
`, rName))
}
