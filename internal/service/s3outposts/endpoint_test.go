// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3outposts_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3outposts "github.com/hashicorp/terraform-provider-aws/internal/service/s3outposts"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3OutpostsEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3outposts_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rInt := sdkacctest.RandIntRange(0, 255)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "s3-outposts", regexache.MustCompile(`outpost/[^/]+/endpoint/[0-9a-z]+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCIDRBlock, "aws_vpc.test", names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct4),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_id", "data.aws_outposts_outpost.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, "aws_subnet.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_type", "Private"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccEndpointImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3OutpostsEndpoint_private(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3outposts_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rInt := sdkacctest.RandIntRange(0, 255)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_private(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "s3-outposts", regexache.MustCompile(`outpost/[^/]+/endpoint/[0-9a-z]+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCIDRBlock, "aws_vpc.test", names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct4),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_id", "data.aws_outposts_outpost.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, "aws_subnet.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_type", "Private"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccEndpointImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3OutpostsEndpoint_customerOwnedIPv4Pool(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3outposts_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rInt := sdkacctest.RandIntRange(0, 255)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_customerOwnedIPv4Pool(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "s3-outposts", regexache.MustCompile(`outpost/[^/]+/endpoint/[0-9a-z]+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCIDRBlock, "aws_vpc.test", names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct4),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_id", "data.aws_outposts_outpost.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, "aws_subnet.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_type", "CustomerOwnedIp"),
					resource.TestMatchResourceAttr(resourceName, "customer_owned_ipv4_pool", regexache.MustCompile(`^ipv4pool-coip-.+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccEndpointImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3OutpostsEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3outposts_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rInt := sdkacctest.RandIntRange(0, 255)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3outposts.ResourceEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3OutpostsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3outposts_endpoint" {
				continue
			}

			_, err := tfs3outposts.FindEndpointByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Outposts Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEndpointExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Outposts Endpoint ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3OutpostsConn(ctx)

		_, err := tfs3outposts.FindEndpointByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccEndpointImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s,%s", rs.Primary.ID, rs.Primary.Attributes["security_group_id"], rs.Primary.Attributes[names.AttrSubnetID]), nil
	}
}

func testAccEndpointConfig_base(rName string, rInt int) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.%[2]d.0.0/16"

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

resource "aws_subnet" "test" {
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  outpost_arn       = data.aws_outposts_outpost.test.arn
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, rInt)
}

func testAccEndpointConfig_basic(rName string, rInt int) string {
	return acctest.ConfigCompose(testAccEndpointConfig_base(rName, rInt), `
resource "aws_s3outposts_endpoint" "test" {
  outpost_id        = data.aws_outposts_outpost.test.id
  security_group_id = aws_security_group.test.id
  subnet_id         = aws_subnet.test.id
}
`)
}

func testAccEndpointConfig_private(rName string, rInt int) string {
	return acctest.ConfigCompose(testAccEndpointConfig_base(rName, rInt), `
resource "aws_s3outposts_endpoint" "test" {
  outpost_id        = data.aws_outposts_outpost.test.id
  security_group_id = aws_security_group.test.id
  subnet_id         = aws_subnet.test.id
  access_type       = "Private"
}
`)
}

func testAccEndpointConfig_customerOwnedIPv4Pool(rName string, rInt int) string {
	return acctest.ConfigCompose(testAccEndpointConfig_base(rName, rInt), fmt.Sprintf(`
data "aws_ec2_coip_pools" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = data.aws_outposts_outpost.test.arn
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_s3outposts_endpoint" "test" {
  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.test]

  outpost_id               = data.aws_outposts_outpost.test.id
  security_group_id        = aws_security_group.test.id
  subnet_id                = aws_subnet.test.id
  access_type              = "CustomerOwnedIp"
  customer_owned_ipv4_pool = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
}
`, rName))
}
