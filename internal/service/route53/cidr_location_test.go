// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53CIDRLocation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_cidr_location.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	locationName := sdkacctest.RandString(16)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCIDRLocationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCIDRLocation_basic(rName, locationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCIDRLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cidr_blocks.*", "200.5.3.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cidr_blocks.*", "200.6.3.0/24"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, locationName),
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

/*

acctest.CheckFrameworkResourceDisappears() cannot currently set top-level list/set/map attributes.

    cidr_location_test.go:55: Step 1/1 error: Check failed: Check 2/2 error: 1 error occurred:
        	* deleting Route 53 CIDR Location (50c328ab-5145-b3ed-77ab-6241355c43fb:wzv44e9s6lr6p7pj)

        InvalidParameter: 1 validation error(s) found.
        - missing required field, ChangeCidrCollectionInput.Changes[0].CidrList.

func TestAccRoute53CIDRLocation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_cidr_location.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	locationName := sdkacctest.RandString(16)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCIDRLocationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCIDRLocation_basic(rName, locationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCIDRLocationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(acctest.Provider, tfroute53.ResourceCIDRLocation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
*/

func TestAccRoute53CIDRLocation_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_cidr_location.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	locationName := sdkacctest.RandString(16)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCIDRLocationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCIDRLocation_basic(rName, locationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCIDRLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cidr_blocks.*", "200.5.3.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cidr_blocks.*", "200.6.3.0/24"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, locationName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCIDRLocation_updated(rName, locationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCIDRLocationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "cidr_blocks.*", "200.5.2.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cidr_blocks.*", "200.6.3.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cidr_blocks.*", "200.6.5.0/24"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, locationName),
				),
			},
		},
	})
}

func testAccCheckCIDRLocationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_cidr_location" {
				continue
			}

			_, err := tfroute53.FindCIDRLocationByTwoPartKey(ctx, conn, rs.Primary.Attributes["cidr_collection_id"], rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route 53 CIDR Location %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCIDRLocationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		_, err := tfroute53.FindCIDRLocationByTwoPartKey(ctx, conn, rs.Primary.Attributes["cidr_collection_id"], rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccCIDRLocation_basic(rName, locationName string) string {
	return fmt.Sprintf(`
resource "aws_route53_cidr_collection" "test" {
  name = %[1]q
}

resource "aws_route53_cidr_location" "test" {
  cidr_collection_id = aws_route53_cidr_collection.test.id
  name               = %[2]q
  cidr_blocks        = ["200.5.3.0/24", "200.6.3.0/24"]
}
`, rName, locationName)
}

func testAccCIDRLocation_updated(rName, locationName string) string {
	return fmt.Sprintf(`
resource "aws_route53_cidr_collection" "test" {
  name = %[1]q
}

resource "aws_route53_cidr_location" "test" {
  cidr_collection_id = aws_route53_cidr_collection.test.id
  name               = %[2]q
  cidr_blocks        = ["200.5.2.0/24", "200.6.3.0/24", "200.6.5.0/24"]
}
`, rName, locationName)
}
