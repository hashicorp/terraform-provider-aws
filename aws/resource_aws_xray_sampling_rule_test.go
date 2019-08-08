package aws

import (
	"fmt"
	"testing"

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
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSXray(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSXraySamplingRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSXraySamplingRuleConfig_basic(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXraySamplingRuleExists(resourceName, &samplingRule),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "xray", fmt.Sprintf("sampling-rule/%s", ruleName)),
					resource.TestCheckResourceAttr(resourceName, "priority", "5"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "reservoir_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "url_path", "*"),
					resource.TestCheckResourceAttr(resourceName, "host", "*"),
					resource.TestCheckResourceAttr(resourceName, "http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "fixed_rate", "0.3"),
					resource.TestCheckResourceAttr(resourceName, "resource_arn", "*"),
					resource.TestCheckResourceAttr(resourceName, "service_name", "*"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "*"),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "1"),
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

func TestAccAWSXraySamplingRule_update(t *testing.T) {
	var samplingRule xray.SamplingRule
	resourceName := "aws_xray_sampling_rule.test"
	rString := acctest.RandString(8)
	ruleName := fmt.Sprintf("tf_acc_sampling_rule_%s", rString)
	updatedPriority := acctest.RandIntRange(0, 9999)
	updatedReservoirSize := acctest.RandIntRange(0, 2147483647)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSXray(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSXraySamplingRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSXraySamplingRuleConfig_update(ruleName, acctest.RandIntRange(0, 9999), acctest.RandIntRange(0, 2147483647)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXraySamplingRuleExists(resourceName, &samplingRule),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "xray", fmt.Sprintf("sampling-rule/%s", ruleName)),
					resource.TestCheckResourceAttrSet(resourceName, "priority"),
					resource.TestCheckResourceAttrSet(resourceName, "reservoir_size"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "url_path", "*"),
					resource.TestCheckResourceAttr(resourceName, "host", "*"),
					resource.TestCheckResourceAttr(resourceName, "http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "fixed_rate", "0.3"),
					resource.TestCheckResourceAttr(resourceName, "resource_arn", "*"),
					resource.TestCheckResourceAttr(resourceName, "service_name", "*"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "*"),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
				),
			},
			{ // Update attributes
				Config: testAccAWSXraySamplingRuleConfig_update(ruleName, updatedPriority, updatedReservoirSize),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXraySamplingRuleExists(resourceName, &samplingRule),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "xray", fmt.Sprintf("sampling-rule/%s", ruleName)),
					resource.TestCheckResourceAttr(resourceName, "priority", fmt.Sprintf("%d", updatedPriority)),
					resource.TestCheckResourceAttr(resourceName, "reservoir_size", fmt.Sprintf("%d", updatedReservoirSize)),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "url_path", "*"),
					resource.TestCheckResourceAttr(resourceName, "host", "*"),
					resource.TestCheckResourceAttr(resourceName, "http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "fixed_rate", "0.3"),
					resource.TestCheckResourceAttr(resourceName, "resource_arn", "*"),
					resource.TestCheckResourceAttr(resourceName, "service_name", "*"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "*"),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
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

func testAccCheckXraySamplingRuleExists(n string, samplingRule *xray.SamplingRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No XRay Sampling Rule ID is set")
		}
		conn := testAccProvider.Meta().(*AWSClient).xrayconn

		rule, err := getXraySamplingRule(conn, rs.Primary.ID)

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

		conn := testAccProvider.Meta().(*AWSClient).xrayconn

		rule, err := getXraySamplingRule(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if rule != nil {
			return fmt.Errorf("Expected XRay Sampling Rule to be destroyed, %s found", rs.Primary.ID)
		}
	}

	return nil
}

func testAccPreCheckAWSXray(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).xrayconn

	input := &xray.GetSamplingRulesInput{}

	_, err := conn.GetSamplingRules(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSXraySamplingRuleConfig_basic(ruleName string) string {
	return fmt.Sprintf(`
resource "aws_xray_sampling_rule" "test" {
  rule_name      = "%s"
  priority       = 5
  reservoir_size = 10
  url_path       = "*"
  host           = "*"
  http_method    = "GET"
  service_type   = "*"
  service_name   = "*"
  fixed_rate     = 0.3
  resource_arn   = "*"
  version        = 1

  attributes = {
    Hello = "World"
  }
}
`, ruleName)
}

func testAccAWSXraySamplingRuleConfig_update(ruleName string, priority int, reservoirSize int) string {
	return fmt.Sprintf(`
resource "aws_xray_sampling_rule" "test" {
  rule_name      = "%s"
  priority       = %d
  reservoir_size = %d
  url_path       = "*"
  host           = "*"
  http_method    = "GET"
  service_type   = "*"
  service_name   = "*"
  fixed_rate     = 0.3
  resource_arn   = "*"
  version        = 1
}
`, ruleName, priority, reservoirSize)
}
