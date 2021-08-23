package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRoute53RecoveryControlConfigSafetyRule_assertionrule(t *testing.T) {
	rClusterName := acctest.RandomWithPrefix("tf-acc-test-cluster")
	rControlPanelName := acctest.RandomWithPrefix("tf-acc-test-control-panel")
	rRoutingControlName := acctest.RandomWithPrefix("tf-acc-test-routing-control")
	rSafetyRuleName := acctest.RandomWithPrefix("tf-acc-test-safety-rule")
	resourceName := "aws_route53recoverycontrolconfig_safety_rule.test.assertion_rule"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53recoverycontrolconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigRoutingControlConfigSafetyRuleAssertion(rClusterName, rControlPanelName, rRoutingControlName, rSafetyRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rSafetyRuleName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "wait_perios_ms", "5000"),
					resource.TestCheckResourceAttr(resourceName, "asserted_controls.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "control_panel_arn", "aws_route53recoverycontrolconfig_control_panel.test", "control_panel_arn"),
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

func TestAccAWSRoute53RecoveryControlConfigSafetyRule_gatingrule(t *testing.T) {
	rClusterName := acctest.RandomWithPrefix("tf-acc-test-cluster")
	rControlPanelName := acctest.RandomWithPrefix("tf-acc-test-control-panel")
	rRoutingControlName := acctest.RandomWithPrefix("tf-acc-test-routing-control")
	rSafetyRuleName := acctest.RandomWithPrefix("tf-acc-test-safety-rule")
	resourceName := "aws_route53recoverycontrolconfig_safety_rule.test.gating_rule"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53recoverycontrolconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigRoutingControlConfigSafetyRuleGating(rClusterName, rControlPanelName, rRoutingControlName, rSafetyRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rSafetyRuleName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "wait_perios_ms", "5000"),
					resource.TestCheckResourceAttr(resourceName, "target_controls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "gating_controls.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "control_panel_arn", "aws_route53recoverycontrolconfig_control_panel.test", "control_panel_arn"),
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

func testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoverycontrolconfig_safety_rule" {
			continue
		}

		input := &route53recoverycontrolconfig.DescribeSafetyRuleInput{
			SafetyRuleArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeSafetyRule(input)

		if err == nil {
			return fmt.Errorf("Route53RecoveryControlConfig Safety Rule (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

		input := &route53recoverycontrolconfig.DescribeSafetyRuleInput{
			SafetyRuleArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeSafetyRule(input)

		return err
	}
}

func testAccAwsRoute53RecoveryControlConfigRoutingControlConfigSafetyRuleAssertion(rName, rName2, rName3, rName4 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_cluster" "test" {
  name = %[1]q
}

resource "aws_route53recoverycontrolconfig_control_panel" "test" {
  name        = %[2]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.cluster_arn
}

resource "aws_route53recoverycontrolconfig_routing_control" "test" {
  name              = %[3]q
  cluster_arn       = aws_route53recoverycontrolconfig_cluster.test.cluster_arn
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test.control_panel_arn
}

resource "aws_route53recoverycontrolconfig_safety_rule" "test" {
  name              = %[4]q
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test.control_panel_arn
  wait_period_ms    = 5000
  asserted_controls = [aws_route53recoverycontrolconfig_routing_control.test.routing_control_arn]
  rule_config       = { inverted = false, threshold = 0, type = "AND"}
}
`, rName, rName2, rName3, rName4)
}

func testAccAwsRoute53RecoveryControlConfigRoutingControlConfigSafetyRuleGating(rName, rName2, rName3, rName4 string) string {
	return fmt.Sprintf(`

resource "aws_route53recoverycontrolconfig_cluster" "test2" {
  name = %[1]q
}

resource "aws_route53recoverycontrolconfig_control_panel" "test2" {
  name        = %[2]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test2.cluster_arn
}

resource "aws_route53recoverycontrolconfig_routing_control" "test2" {
  name              = %[3]q
  cluster_arn       = aws_route53recoverycontrolconfig_cluster.test.cluster_arn
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test2.control_panel_arn
}

resource "aws_route53recoverycontrolconfig_safety_rule" "test2" {
  name              = %[4]q
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test2.control_panel_arn
  wait_period_ms    = 5000
  gating_controls   = [aws_route53recoverycontrolconfig_routing_control.test2.routing_control_arn]
  target_controls   = [aws_route53recoverycontrolconfig_routing_control.test2.routing_control_arn]
  rule_config       = { inverted = false, threshold = 0, type = "AND"}
}
`, rName, rName2, rName3, rName4)
}