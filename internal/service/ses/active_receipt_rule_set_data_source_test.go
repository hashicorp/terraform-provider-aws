// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/ses"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccActiveReceiptRuleSetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "data.aws_ses_active_receipt_rule_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActiveReceiptRuleSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActiveReceiptRuleSetDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveReceiptRuleSetExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "ses", fmt.Sprintf("receipt-rule-set/%s", rName)),
				),
			},
		},
	})
}

func testAccActiveReceiptRuleSetDataSource_noActiveRuleSet(t *testing.T) {
	ctx := acctest.Context(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckUnsetActiveRuleSet(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccActiveReceiptRuleSetDataSourceConfig_noActiveRuleSet(),
				ExpectError: regexache.MustCompile("empty result"),
			},
		},
	})
}

func testAccActiveReceiptRuleSetDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_ses_active_receipt_rule_set" "test" {
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
}

data "aws_ses_active_receipt_rule_set" "test" {
  depends_on = [aws_ses_active_receipt_rule_set.test]
}
`, name)
}

func testAccActiveReceiptRuleSetDataSourceConfig_noActiveRuleSet() string {
	return `
data "aws_ses_active_receipt_rule_set" "test" {}
`
}

func testAccPreCheckUnsetActiveRuleSet(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

	output, err := conn.DescribeActiveReceiptRuleSetWithContext(ctx, &ses.DescribeActiveReceiptRuleSetInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if output == nil || output.Metadata == nil {
		return
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	_, err = conn.SetActiveReceiptRuleSetWithContext(ctx, &ses.SetActiveReceiptRuleSetInput{
		RuleSetName: nil,
	})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
