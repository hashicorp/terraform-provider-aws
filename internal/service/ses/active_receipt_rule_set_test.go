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

// Only one SES Receipt RuleSet can be active at a time, so run serially
// locally and in TeamCity.
func TestAccSESActiveReceiptRuleSet_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Resource": {
			acctest.CtBasic:      testAccActiveReceiptRuleSet_basic,
			acctest.CtDisappears: testAccActiveReceiptRuleSet_disappears,
		},
		"DataSource": {
			acctest.CtBasic:   testAccActiveReceiptRuleSetDataSource_basic,
			"noActiveRuleSet": testAccActiveReceiptRuleSetDataSource_noActiveRuleSet,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccActiveReceiptRuleSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_active_receipt_rule_set.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActiveReceiptRuleSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccActiveReceiptRuleSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveReceiptRuleSetExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", fmt.Sprintf("receipt-rule-set/%s", rName)),
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

func testAccActiveReceiptRuleSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_active_receipt_rule_set.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActiveReceiptRuleSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccActiveReceiptRuleSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveReceiptRuleSetExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfses.ResourceActiveReceiptRuleSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckActiveReceiptRuleSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_active_receipt_rule_set" {
				continue
			}

			_, err := tfses.FindActiveReceiptRuleSet(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SES Active Receipt Rule Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckActiveReceiptRuleSetExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		_, err := tfses.FindActiveReceiptRuleSet(ctx, conn)

		return err
	}
}

func testAccActiveReceiptRuleSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_ses_active_receipt_rule_set" "test" {
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
}
`, name)
}
