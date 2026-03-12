// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESReceiptFilter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ses_receipt_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckReceiptRule(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptFilterExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", fmt.Sprintf("receipt-filter/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cidr", "10.10.10.10"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, "Block"),
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

func TestAccSESReceiptFilter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ses_receipt_filter.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckReceiptRule(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptFilterExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfses.ResourceReceiptFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReceiptFilterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_receipt_filter" {
				continue
			}

			_, err := tfses.FindReceiptFilterByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SES Receipt Filter %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckReceiptFilterExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		_, err := tfses.FindReceiptFilterByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccReceiptFilterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_filter" "test" {
  cidr   = "10.10.10.10"
  name   = %[1]q
  policy = "Block"
}
`, rName)
}
