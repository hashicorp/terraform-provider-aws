// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMPool_basic(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	var pool awstypes.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "allocation_default_netmask_length"),
					resource.TestCheckNoResourceAttr(resourceName, "allocation_max_netmask_length"),
					resource.TestCheckNoResourceAttr(resourceName, "allocation_min_netmask_length"),
					resource.TestCheckResourceAttr(resourceName, "allocation_resource_tags.%", "0"),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "ec2", "ipam-pool/{id}"),
					resource.TestCheckResourceAttr(resourceName, "auto_import", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "aws_service", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "ipam_scope_type"),
					resource.TestCheckResourceAttr(resourceName, "locale", "None"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_depth"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "create-complete"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMPoolConfig_updated,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "allocation_default_netmask_length", "32"),
					resource.TestCheckResourceAttr(resourceName, "allocation_max_netmask_length", "32"),
					resource.TestCheckResourceAttr(resourceName, "allocation_min_netmask_length", "32"),
					resource.TestCheckResourceAttr(resourceName, "allocation_resource_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "allocation_resource_tags.test", "1"),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "ec2", "ipam-pool/{id}"),
					resource.TestCheckResourceAttr(resourceName, "auto_import", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "aws_service", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrSet(resourceName, "ipam_scope_type"),
					resource.TestCheckResourceAttr(resourceName, "locale", "None"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_depth"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "modify-complete"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccIPAMPool_disappears(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	var pool awstypes.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceIPAMPool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIPAMPool_ipv6Basic(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	var pool awstypes.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_ipv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "publicly_advertisable", acctest.CtFalse),
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

func TestAccIPAMPool_ipv6PublicIPAmazon(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	var pool awstypes.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_ipv6PublicIPAmazon,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "public_ip_source", "amazon"),
					resource.TestCheckResourceAttr(resourceName, "aws_service", "ec2"),
					resource.TestCheckResourceAttr(resourceName, "publicly_advertisable", acctest.CtFalse),
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

func TestAccIPAMPool_ipv6Contiguous(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	var pool awstypes.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_ipv6Contiguous,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "public_ip_source", "byoip"),
					resource.TestCheckResourceAttr(resourceName, "publicly_advertisable", acctest.CtFalse),
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

func TestAccIPAMPool_cascade(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	var pool awstypes.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_cascade,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					testAccCheckIPAMPoolCIDRCreate(ctx, &pool),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cascade"},
			},
		},
	})
}

func TestAccIPAMPool_tags(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	var pool awstypes.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_tags(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMPoolConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccIPAMPoolConfig_tags(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccIPAMPool_ipv6PrivateScope(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	var pool awstypes.IpamPool
	resourceName := "aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolConfig_ipv6PrivateScope,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMPoolExists(ctx, resourceName, &pool),
				),
			},
		},
	})
}

func testAccCheckIPAMPoolExists(ctx context.Context, n string, v *awstypes.IpamPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindIPAMPoolByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIPAMPoolDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_ipam_pool" {
				continue
			}

			_, err := tfec2.FindIPAMPoolByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IPAM Pool still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIPAMPoolCIDRCreate(ctx context.Context, ipampool *awstypes.IpamPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		input := ec2.ProvisionIpamPoolCidrInput{
			IpamPoolId: ipampool.IpamPoolId,
			Cidr:       aws.String("10.0.0.0/16"),
		}
		_, err := conn.ProvisionIpamPoolCidr(ctx, &input)

		return err
	}
}

const testAccIPAMPoolConfig_base = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

var testAccIPAMPoolConfig_basic = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
}
`)

var testAccIPAMPoolConfig_cascade = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  cascade        = true
}
`)

var testAccIPAMPoolConfig_updated = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.private_default_scope_id
  auto_import                       = true
  allocation_default_netmask_length = 32
  allocation_max_netmask_length     = 32
  allocation_min_netmask_length     = 32
  allocation_resource_tags = {
    test = "1"
  }
  description = "test"
}
`)

var testAccIPAMPoolConfig_ipv6 = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family        = "ipv6"
  ipam_scope_id         = aws_vpc_ipam.test.public_default_scope_id
  locale                = data.aws_region.current.name
  publicly_advertisable = false
}
`)

var testAccIPAMPoolConfig_ipv6PublicIPAmazon = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family   = "ipv6"
  ipam_scope_id    = aws_vpc_ipam.test.public_default_scope_id
  locale           = data.aws_region.current.name
  public_ip_source = "amazon"
  aws_service      = "ec2"
}
`)

var testAccIPAMPoolConfig_ipv6Contiguous = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family        = "ipv6"
  ipam_scope_id         = aws_vpc_ipam.test.public_default_scope_id
  locale                = data.aws_region.current.name
  public_ip_source      = "byoip"
  aws_service           = "ec2"
  publicly_advertisable = false
}
`)

func testAccIPAMPoolConfig_tags(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccIPAMPoolConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccIPAMPoolConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccIPAMPoolConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

var testAccIPAMPoolConfig_ipv6PrivateScope = acctest.ConfigCompose(testAccIPAMScopeConfig_base, `
resource "aws_vpc_ipam_scope" "test" {
  ipam_id = aws_vpc_ipam.test.id
}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv6"
  ipam_scope_id  = aws_vpc_ipam_scope.test.id
  locale         = data.aws_region.current.name
}
`)
