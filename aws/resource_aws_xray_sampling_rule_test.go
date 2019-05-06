package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSXraySamplingRule_basic(t *testing.T) {
	var samplingRule xray.SamplingRule
	resourceName := "aws_xray_sampling_rule.test"
	rString := acctest.RandString(8)
	ruleName := fmt.Sprintf("tf_acc_sampling_rule_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSXraySamplingRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSXraySamplingRule_basic(ruleName, acctest.RandIntRange(0, 9999), acctest.RandIntRange(0, 2147483647)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXraySamplingRuleExists(resourceName, &samplingRule),
					testAccCheckResourceAttrRegionalARN(resourceName, "rule_arn", "xray", fmt.Sprintf("sampling-rule/%s", ruleName)),
				),
			},
		},
	})
}

func testAccAWSXraySamplingRule_basic(ruleName string, priority int, reservoirSize int) string {
	return fmt.Sprintf(`
resource "aws_xray_sampling_rule" "test" {
	rule_name = "%s"
	priority = %d
	version = 1
	reservoir_size = %d
	url_path = "*"
	host = "*"
	http_method = "GET"
	service_type = "*"
	service_name = "*"
	fixed_rate = 0.3
	resource_arn = "*"
}
`, ruleName, priority, reservoirSize)
}

func testAccCheckXraySamplingRuleExists(n string, samplingRule *xray.SamplingRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No XRay Sampling Rule ID is set")
		}

		rule, err := getSamplingRule(rs.Primary.ID)

		if err != nil {
			return err
		}

		*samplingRule = *rule

		return nil
	}
}

func testAccCheckAWSXraySamplingRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_xray_sampling_rule" {
			continue
		}

		_, err := getSamplingRule(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Expected XRay Sampling Rule to be destroyed, %s found", rs.Primary.ID)
		}
	}

	return nil
}

func getSamplingRule(ruleName string) (*xray.SamplingRule, error) {
	conn := testAccProvider.Meta().(*AWSClient).xrayconn
	params := &xray.GetSamplingRulesInput{}
	for {
		out, err := conn.GetSamplingRules(params)
		if err != nil {
			return nil, err
		}
		for _, samplingRuleRecord := range out.SamplingRuleRecords {
			samplingRule := samplingRuleRecord.SamplingRule
			if aws.StringValue(samplingRule.RuleName) == ruleName {
				return samplingRule, nil
			}
		}
		if out.NextToken == nil {
			break
		}
		params.NextToken = out.NextToken
	}
	return nil, fmt.Errorf("XRay Sampling Rule: %s not found found", ruleName)
}
