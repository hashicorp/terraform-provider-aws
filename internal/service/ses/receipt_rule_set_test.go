package ses_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_ses_receipt_rule_set", &resource.Sweeper{
		Name: "aws_ses_receipt_rule_set",
		F:    sweepReceiptRuleSets,
	})
}

func sweepReceiptRuleSets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SESConn

	// You cannot delete the receipt rule set that is currently active.
	// Setting the name of the receipt rule set to make active to null disables all email receiving.
	log.Printf("[INFO] Disabling any currently active SES Receipt Rule Set")
	_, err = conn.SetActiveReceiptRuleSet(&ses.SetActiveReceiptRuleSetInput{})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SES Receipt Rule Sets sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error disabling any currently active SES Receipt Rule Set: %w", err)
	}

	input := &ses.ListReceiptRuleSetsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListReceiptRuleSets(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SES Receipt Rule Sets sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SES Receipt Rule Sets: %w", err))
			return sweeperErrs
		}

		for _, ruleSet := range output.RuleSets {
			name := aws.StringValue(ruleSet.Name)

			log.Printf("[INFO] Deleting SES Receipt Rule Set: %s", name)
			_, err := conn.DeleteReceiptRuleSet(&ses.DeleteReceiptRuleSetInput{
				RuleSetName: aws.String(name),
			})
			if tfawserr.ErrMessageContains(err, ses.ErrCodeRuleSetDoesNotExistException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting SES Receipt Rule Set (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccSESReceiptRuleSet_basic(t *testing.T) {
	resourceName := "aws_ses_receipt_rule_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t); testAccPreCheckSESReceiptRule(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSESReceiptRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_set_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("receipt-rule-set/%s", rName)),
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

func TestAccSESReceiptRuleSet_disappears(t *testing.T) {
	resourceName := "aws_ses_receipt_rule_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t); testAccPreCheckSESReceiptRule(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSESReceiptRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReceiptRuleSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReceiptRuleSetExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfses.ResourceReceiptRuleSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSESReceiptRuleSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_receipt_rule_set" {
			continue
		}

		params := &ses.DescribeReceiptRuleSetInput{
			RuleSetName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeReceiptRuleSet(params)

		if tfawserr.ErrMessageContains(err, ses.ErrCodeRuleSetDoesNotExistException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SES Receipt Rule Set (%s) still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckReceiptRuleSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Receipt Rule Set not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Receipt Rule Set name not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		params := &ses.DescribeReceiptRuleSetInput{
			RuleSetName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeReceiptRuleSet(params)
		return err
	}
}

func testAccReceiptRuleSetConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %q
}
`, rName)
}
