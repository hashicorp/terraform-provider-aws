// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIPAMScope_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var scope ec2.IpamScope
	resourceName := "aws_vpc_ipam_scope.test"
	ipamName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMScopeConfig_basic("test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_arn", ipamName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_id", ipamName, "id"),
					resource.TestCheckResourceAttr(resourceName, "is_default", "false"),
					resource.TestCheckResourceAttr(resourceName, "pool_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "description", "test2"),
				),
			},
		},
	})
}

func TestAccIPAMScope_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var scope ec2.IpamScope
	resourceName := "aws_vpc_ipam_scope.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
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
	var scope ec2.IpamScope
	resourceName := "aws_vpc_ipam_scope.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPAMScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMScopeConfig_tags("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
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
				Config: testAccIPAMScopeConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIPAMScopeConfig_tags("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPAMScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckIPAMScopeExists(ctx context.Context, n string, v *ec2.IpamScope) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPAM Scope ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

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
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

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
