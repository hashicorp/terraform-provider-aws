// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53CIDRCollection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CollectionSummary
	resourceName := "aws_route53_cidr_collection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCIDRCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCIDRCollection_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCIDRCollectionExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrGlobalARNNoAccount(resourceName, names.AttrARN, "route53", regexache.MustCompile(`cidrcollection/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccRoute53CIDRCollection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CollectionSummary
	resourceName := "aws_route53_cidr_collection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCIDRCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCIDRCollection_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCIDRCollectionExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceCIDRCollection, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCIDRCollectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_cidr_collection" {
				continue
			}

			_, err := tfroute53.FindCIDRCollectionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route 53 CIDR Collection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCIDRCollectionExists(ctx context.Context, n string, v *awstypes.CollectionSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		output, err := tfroute53.FindCIDRCollectionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCIDRCollection_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_cidr_collection" "test" {
  name = %[1]q
}
`, rName)
}
