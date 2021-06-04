package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/atest"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

func init() {
	resource.AddTestSweepers("aws_ses_receipt_rule_set", &resource.Sweeper{
		Name: "aws_ses_receipt_rule_set",
		F:    testSweepSesReceiptRuleSets,
	})
}

func testSweepSesReceiptRuleSets(region string) error {
	client, err := atest.SharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*awsprovider.AWSClient).SESConn

	// You cannot delete the receipt rule set that is currently active.
	// Setting the name of the receipt rule set to make active to null disables all email receiving.
	log.Printf("[INFO] Disabling any currently active SES Receipt Rule Set")
	_, err = conn.SetActiveReceiptRuleSet(&ses.SetActiveReceiptRuleSetInput{})
	if atest.SweepSkipSweepError(err) {
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
		if atest.SweepSkipSweepError(err) {
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

func TestAccAWSSESReceiptRuleSet_basic(t *testing.T) {
	resourceName := "aws_ses_receipt_rule_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			atest.PreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   atest.ErrorCheck(t, ses.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckSESReceiptRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_set_name", rName),
					atest.CheckAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("receipt-rule-set/%s", rName)),
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

func TestAccAWSSESReceiptRuleSet_disappears(t *testing.T) {
	resourceName := "aws_ses_receipt_rule_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			atest.PreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   atest.ErrorCheck(t, ses.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckSESReceiptRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleSetExists(resourceName),
					atest.CheckDisappears(atest.Provider, resourceAwsSesReceiptRuleSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSESReceiptRuleSetDestroy(s *terraform.State) error {
	conn := atest.Provider.Meta().(*awsprovider.AWSClient).SESConn

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

func testAccCheckAwsSESReceiptRuleSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Receipt Rule Set not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Receipt Rule Set name not set")
		}

		conn := atest.Provider.Meta().(*awsprovider.AWSClient).SESConn

		params := &ses.DescribeReceiptRuleSetInput{
			RuleSetName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeReceiptRuleSet(params)
		return err
	}
}

func testAccAWSSESReceiptRuleSetConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %q
}
`, rName)
}
