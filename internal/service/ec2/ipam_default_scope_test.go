// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMDefaultScope_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var scope awstypes.IpamScope
	resourceName := "aws_vpc_ipam_default_scope.test"
	ipamName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMDefaultScopeConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, ipamName, "private_default_scope_id"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_arn", ipamName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_id", ipamName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "is_default", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "pool_count", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"default_scope_id"},
			},
		},
	})
}

// It is not possible to delete the default scope so we'll test parent IPAM disappears instead
func TestAccIPAMDefaultScope_disappears_IPAM(t *testing.T) {
	ctx := acctest.Context(t)
	var scope awstypes.IpamScope
	parentResourceName := "aws_vpc_ipam.test"
	resourceName := "aws_vpc_ipam_default_scope.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMDefaultScopeConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceIPAM(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIPAMDefaultScope_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var scope awstypes.IpamScope
	resourceName := "aws_vpc_ipam_default_scope.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMDefaultScopeConfig_tags(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"default_scope_id"},
			},
			{
				Config: testAccIPAMDefaultScopeConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccIPAMDefaultScopeConfig_tags(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

const testAccIPAMDefaultScopeConfig_base = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

func testAccIPAMDefaultScopeConfig_basic() string {
	return acctest.ConfigCompose(testAccIPAMDefaultScopeConfig_base, `
resource "aws_vpc_ipam_default_scope" "test" {
  default_scope_id = aws_vpc_ipam.test.private_default_scope_id
}
`)
}

func testAccIPAMDefaultScopeConfig_tags(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccIPAMDefaultScopeConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_default_scope" "test" {
  default_scope_id = aws_vpc_ipam.test.private_default_scope_id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccIPAMDefaultScopeConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccIPAMDefaultScopeConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_default_scope" "test" {
  default_scope_id = aws_vpc_ipam.test.private_default_scope_id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
