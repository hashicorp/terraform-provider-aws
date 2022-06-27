package ses_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
)

// Only one SES Receipt RuleSet can be active at a time, so run serially
// locally and in TeamCity.
func TestAccSESActiveReceiptRuleSet_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Resource": {
			"basic":      testAccActiveReceiptRuleSet_basic,
			"disappears": testAccActiveReceiptRuleSet_disappears,
		},
		"DataSource": {
			"basic": testAccActiveReceiptRuleSetDataSource_basic,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccActiveReceiptRuleSet_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_active_receipt_rule_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckActiveReceiptRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccActiveReceiptRuleSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveReceiptRuleSetExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("receipt-rule-set/%s", rName)),
				),
			},
		},
	})
}

func testAccActiveReceiptRuleSet_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_active_receipt_rule_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckReceiptRule(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckActiveReceiptRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccActiveReceiptRuleSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveReceiptRuleSetExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfses.ResourceActiveReceiptRuleSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckActiveReceiptRuleSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_active_receipt_rule_set" {
			continue
		}

		response, err := conn.DescribeActiveReceiptRuleSet(&ses.DescribeActiveReceiptRuleSetInput{})
		if err != nil {
			return err
		}

		if response.Metadata != nil && (aws.StringValue(response.Metadata.Name) == rs.Primary.ID) {
			return fmt.Errorf("Active receipt rule set still exists")
		}

	}

	return nil

}

func testAccCheckActiveReceiptRuleSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Active Receipt Rule Set not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Active Receipt Rule Set name not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		response, err := conn.DescribeActiveReceiptRuleSet(&ses.DescribeActiveReceiptRuleSetInput{})
		if err != nil {
			return err
		}

		if response.Metadata != nil && (aws.StringValue(response.Metadata.Name) != rs.Primary.ID) {
			return fmt.Errorf("The active receipt rule set (%s) was not set to %s", aws.StringValue(response.Metadata.Name), rs.Primary.ID)
		}

		return nil
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
