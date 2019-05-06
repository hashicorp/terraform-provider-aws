package aws

import (
	"fmt"
	"regexp"
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
				Config: testAccAWSXraySamplingRuleConfig_basic(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXraySamplingRuleExists(resourceName, &samplingRule),
					testAccCheckResourceAttrRegionalARN(resourceName, "rule_arn", "xray", fmt.Sprintf("sampling-rule/%s", ruleName)),
					resource.TestCheckResourceAttrSet(resourceName, "priority"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "reservoir_size"),
					resource.TestCheckResourceAttrSet(resourceName, "url_path"),
					resource.TestCheckResourceAttrSet(resourceName, "host"),
					resource.TestCheckResourceAttrSet(resourceName, "http_method"),
					resource.TestCheckResourceAttrSet(resourceName, "fixed_rate"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "service_name"),
					resource.TestCheckResourceAttrSet(resourceName, "service_type"),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "1"),
				),
			},
		},
	})
}

func TestAccAWSXraySamplingRule_update(t *testing.T) {
	var samplingRule xray.SamplingRule
	resourceName := "aws_xray_sampling_rule.test"
	rString := acctest.RandString(8)
	ruleName := fmt.Sprintf("tf_acc_sampling_rule_%s", rString)
	initialVersion := acctest.RandIntRange(1, 3)
	updatedVersion := initialVersion + 1
	updatedPriority := acctest.RandIntRange(0, 9999)
	updatedReservoirSize := acctest.RandIntRange(0, 2147483647)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSXraySamplingRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSXraySamplingRuleConfig_update(ruleName, acctest.RandIntRange(0, 9999), acctest.RandIntRange(0, 2147483647), initialVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXraySamplingRuleExists(resourceName, &samplingRule),
					testAccCheckResourceAttrRegionalARN(resourceName, "rule_arn", "xray", fmt.Sprintf("sampling-rule/%s", ruleName)),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
				),
			},
			{ // Update attributes
				Config: testAccAWSXraySamplingRuleConfig_update(ruleName, updatedPriority, updatedReservoirSize, initialVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXraySamplingRuleExists(resourceName, &samplingRule),
					testAccCheckResourceAttrRegionalARN(resourceName, "rule_arn", "xray", fmt.Sprintf("sampling-rule/%s", ruleName)),
					resource.TestCheckResourceAttr(resourceName, "priority", fmt.Sprintf("%d", updatedPriority)),
					resource.TestCheckResourceAttr(resourceName, "reservoir_size", fmt.Sprintf("%d", updatedReservoirSize)),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
				),
			},
			{ // Increment version
				Config:      testAccAWSXraySamplingRuleConfig_update(ruleName, updatedPriority, updatedReservoirSize, updatedVersion),
				ExpectError: regexp.MustCompile(`Version cannot be modified`),
			},
		},
	})
}

func TestAccAWSXraySamplingRule_import(t *testing.T) {
	resourceName := "aws_xray_sampling_rule.test"
	rString := acctest.RandString(8)
	ruleName := fmt.Sprintf("tf_acc_sampling_rule_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSXraySamplingRuleConfig_basic(ruleName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
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

func testAccAWSXraySamplingRuleConfig_basic(ruleName string) string {
	return fmt.Sprintf(`
resource "aws_xray_sampling_rule" "test" {
	rule_name = "%s"
	priority = 5
	reservoir_size = 10
	url_path = "*"
	host = "*"
	http_method = "GET"
	service_type = "*"
	service_name = "*"
	fixed_rate = 0.3
	resource_arn = "*"
	version = 1
	attributes = {
		Hello = "World"
	}
}
`, ruleName)
}

func testAccAWSXraySamplingRuleConfig_update(ruleName string, priority int, reservoirSize int, version int) string {
	return fmt.Sprintf(`
resource "aws_xray_sampling_rule" "test" {
	rule_name = "%s"
	priority = %d
	reservoir_size = %d
	url_path = "*"
	host = "*"
	http_method = "GET"
	service_type = "*"
	service_name = "*"
	fixed_rate = 0.3
	resource_arn = "*"
	version = %d
}
`, ruleName, priority, reservoirSize, version)
}
