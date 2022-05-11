package xray_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/xray"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfxray "github.com/hashicorp/terraform-provider-aws/internal/service/xray"
)

func TestAccXRaySamplingRule_basic(t *testing.T) {
	var samplingRule xray.SamplingRule
	resourceName := "aws_xray_sampling_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, xray.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSamplingRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSamplingRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamplingRuleExists(resourceName, &samplingRule),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "xray", fmt.Sprintf("sampling-rule/%s", rName)),
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
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccXRaySamplingRule_update(t *testing.T) {
	var samplingRule xray.SamplingRule
	resourceName := "aws_xray_sampling_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedPriority := sdkacctest.RandIntRange(0, 9999)
	updatedReservoirSize := sdkacctest.RandIntRange(0, 2147483647)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, xray.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSamplingRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSamplingRuleConfig_update(rName, sdkacctest.RandIntRange(0, 9999), sdkacctest.RandIntRange(0, 2147483647)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamplingRuleExists(resourceName, &samplingRule),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "xray", fmt.Sprintf("sampling-rule/%s", rName)),
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
				Config: testAccSamplingRuleConfig_update(rName, updatedPriority, updatedReservoirSize),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamplingRuleExists(resourceName, &samplingRule),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "xray", fmt.Sprintf("sampling-rule/%s", rName)),
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

func TestAccXRaySamplingRule_tags(t *testing.T) {
	var samplingRule xray.SamplingRule
	resourceName := "aws_xray_sampling_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, xray.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSamplingRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSamplingRuleTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamplingRuleExists(resourceName, &samplingRule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSamplingRuleTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamplingRuleExists(resourceName, &samplingRule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSamplingRuleTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamplingRuleExists(resourceName, &samplingRule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccXRaySamplingRule_disappears(t *testing.T) {
	var samplingRule xray.SamplingRule
	resourceName := "aws_xray_sampling_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, xray.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSamplingRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSamplingRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamplingRuleExists(resourceName, &samplingRule),
					acctest.CheckResourceDisappears(acctest.Provider, tfxray.ResourceSamplingRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSamplingRuleExists(n string, samplingRule *xray.SamplingRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No XRay Sampling Rule ID is set")
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).XRayConn

		rule, err := tfxray.GetSamplingRule(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*samplingRule = *rule

		return nil
	}
}

func testAccCheckSamplingRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_xray_sampling_rule" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).XRayConn

		rule, err := tfxray.GetSamplingRule(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if rule != nil {
			return fmt.Errorf("Expected XRay Sampling Rule to be destroyed, %s found", rs.Primary.ID)
		}
	}

	return nil
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).XRayConn

	input := &xray.GetSamplingRulesInput{}

	_, err := conn.GetSamplingRules(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSamplingRuleConfig_basic(ruleName string) string {
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

func testAccSamplingRuleConfig_update(ruleName string, priority int, reservoirSize int) string {
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

func testAccSamplingRuleTags1Config(ruleName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_xray_sampling_rule" "test" {
  rule_name      = %[1]q
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

  tags = {
    %[2]q = %[3]q
  }
}
`, ruleName, tagKey1, tagValue1)
}

func testAccSamplingRuleTags2Config(ruleName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_xray_sampling_rule" "test" {
  rule_name      = %[1]q
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, ruleName, tagKey1, tagValue1, tagKey2, tagValue2)
}
