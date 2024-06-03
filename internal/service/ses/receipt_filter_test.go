// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESReceiptFilter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ses_receipt_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckReceiptRule(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptFilterExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "ses", fmt.Sprintf("receipt-filter/%s", rName)),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckReceiptRule(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReceiptFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptFilterExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfses.ResourceReceiptFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReceiptFilterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_receipt_filter" {
				continue
			}

			response, err := conn.ListReceiptFiltersWithContext(ctx, &ses.ListReceiptFiltersInput{})
			if err != nil {
				return err
			}

			for _, element := range response.Filters {
				if aws.StringValue(element.Name) == rs.Primary.ID {
					return fmt.Errorf("SES Receipt Filter (%s) still exists", rs.Primary.ID)
				}
			}
		}

		return nil
	}
}

func testAccCheckReceiptFilterExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES receipt filter not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES receipt filter ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		response, err := conn.ListReceiptFiltersWithContext(ctx, &ses.ListReceiptFiltersInput{})
		if err != nil {
			return err
		}

		for _, element := range response.Filters {
			if aws.StringValue(element.Name) == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("The receipt filter was not created")
	}
}

func testAccReceiptFilterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_filter" "test" {
  cidr   = "10.10.10.10"
  name   = %q
  policy = "Block"
}
`, rName)
}
