package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func testAccAWSRoute53RecoveryControlConfigSafetyRule_assertionRule(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoverycontrolconfig_safety_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(r53rcc.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, r53rcc.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigRoutingControlConfigSafetyRuleAssertion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "wait_period_ms", "5000"),
					resource.TestCheckResourceAttr(resourceName, "asserted_controls.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "control_panel_arn", "aws_route53recoverycontrolconfig_control_panel.test", "arn"),
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

func testAccAWSRoute53RecoveryControlConfigSafetyRule_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoverycontrolconfig_safety_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(r53rcc.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, r53rcc.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigRoutingControlConfigSafetyRuleAssertion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceSafetyRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSRoute53RecoveryControlConfigSafetyRule_gatingRule(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoverycontrolconfig_safety_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(r53rcc.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, r53rcc.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigRoutingControlConfigSafetyRuleGating(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigSafetyRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "wait_period_ms", "5000"),
					resource.TestCheckResourceAttr(resourceName, "target_controls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "gating_controls.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "control_panel_arn", "aws_route53recoverycontrolconfig_control_panel.test", "arn"),
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryControlConfigConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoverycontrolconfig_safety_rule" {
			continue
		}

		input := &r53rcc.DescribeSafetyRuleInput{
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryControlConfigConn

		input := &r53rcc.DescribeSafetyRuleInput{
			SafetyRuleArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeSafetyRule(input)

		return err
	}
}

func testAccAwsRoute53RecoveryControlConfigRoutingControlConfigSafetyRuleAssertion(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_cluster" "test" {
  name = %[1]q
}

resource "aws_route53recoverycontrolconfig_control_panel" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}

resource "aws_route53recoverycontrolconfig_routing_control" "test" {
  name              = %[1]q
  cluster_arn       = aws_route53recoverycontrolconfig_cluster.test.arn
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test.arn
}

resource "aws_route53recoverycontrolconfig_safety_rule" "test" {
  name              = %[1]q
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test.arn
  wait_period_ms    = 5000
  asserted_controls = [aws_route53recoverycontrolconfig_routing_control.test.arn]

  rule_config {
    inverted  = false
    threshold = 0
    type      = "AND"
  }
}
`, rName)
}

func testAccAwsRoute53RecoveryControlConfigRoutingControlConfigSafetyRuleGating(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_cluster" "test" {
  name = %[1]q
}

resource "aws_route53recoverycontrolconfig_control_panel" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}

resource "aws_route53recoverycontrolconfig_routing_control" "test" {
  name              = %[1]q
  cluster_arn       = aws_route53recoverycontrolconfig_cluster.test.arn
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test.arn
}

resource "aws_route53recoverycontrolconfig_safety_rule" "test" {
  name              = %[1]q
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test.arn
  wait_period_ms    = 5000
  gating_controls   = [aws_route53recoverycontrolconfig_routing_control.test.arn]
  target_controls   = [aws_route53recoverycontrolconfig_routing_control.test.arn]

  rule_config {
    inverted  = false
    threshold = 0
    type      = "AND"
  }
}
`, rName)
}
