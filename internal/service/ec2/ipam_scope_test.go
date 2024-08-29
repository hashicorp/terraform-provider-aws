// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMScope_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var scope awstypes.IpamScope
	resourceName := "aws_vpc_ipam_scope.test"
	ipamName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMScopeConfig_basic("test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_arn", ipamName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_id", ipamName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "is_default", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "pool_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIPAMScopeConfig_basic("test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test2"),
				),
			},
		},
	})
}

func TestAccIPAMScope_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var scope awstypes.IpamScope
	resourceName := "aws_vpc_ipam_scope.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMScopeConfig_basic("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceIPAMScope(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIPAMScope_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var scope awstypes.IpamScope
	resourceName := "aws_vpc_ipam_scope.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMScopeConfig_tags(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
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
				Config: testAccIPAMScopeConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccIPAMScopeConfig_tags(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckIPAMScopeExists(ctx context.Context, n string, v *awstypes.IpamScope) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPAM Scope ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindIPAMScopeByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIPAMScopeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_ipam_scope" {
				continue
			}

			_, err := tfec2.FindIPAMScopeByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IPAM Scope still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

const testAccIPAMScopeConfig_base = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
`

func testAccIPAMScopeConfig_basic(description string) string {
	return acctest.ConfigCompose(testAccIPAMScopeConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_scope" "test" {
  ipam_id     = aws_vpc_ipam.test.id
  description = %[1]q
}
`, description))
}

func testAccIPAMScopeConfig_tags(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccIPAMScopeConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_scope" "test" {
  ipam_id = aws_vpc_ipam.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccIPAMScopeConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccIPAMScopeConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_scope" "test" {
  ipam_id = aws_vpc_ipam.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
