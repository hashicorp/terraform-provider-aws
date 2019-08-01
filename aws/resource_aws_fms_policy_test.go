package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSFmsPolicy_importBasic(t *testing.T) {
	resourceName := "aws_fms_policy.test"
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig(fmsPolicyName, wafRuleGroupName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSFmsPolicy_basic(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					testAccMatchResourceAttrRegionalARN("aws_fms_policy.test", "arn", "fms", regexp.MustCompile(`policy/`)),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "security_service_policy_data.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSFmsPolicy_tags(t *testing.T) {
	fmsPolicyName := fmt.Sprintf("tf-fms-%s", acctest.RandString(5))
	wafRuleGroupName := fmt.Sprintf("tf-waf-rg-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsFmsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsPolicyConfig_tags(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.%", "2"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.Usage", "original"),
				),
			},
			{
				Config: testAccFmsPolicyConfig_tagsChanged(fmsPolicyName, wafRuleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsFmsPolicyExists("aws_fms_policy.test"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "name", fmsPolicyName),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.%", "1"),
					resource.TestCheckResourceAttr("aws_fms_policy.test", "resource_tags.Usage", "changed"),
				),
			},
		},
	})
}

func testAccCheckAwsFmsPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).fmsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fms_policy" {
			continue
		}

		policyId := rs.Primary.Attributes["id"]

		input := &fms.GetPolicyInput{
			PolicyId: aws.String(policyId),
		}

		resp, err := conn.GetPolicy(input)

		if isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp.Policy.PolicyId != nil {
			return fmt.Errorf("[DESTROY Error] Fms Policy (%s) not deleted", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAwsFmsPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccFmsPolicyConfig(name string, group string) string {
	return fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type = "WAF"

    managed_service_data {
      type = "WAF"

      rule_groups {
        id = "${aws_wafregional_rule_group.test.id}"

        override_action={
          type = "COUNT"
        }
      }

      default_action={
        type = "BLOCK"
      }
    }
  }
}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = %[2]q
}
`, name, group)
}

func testAccFmsPolicyConfig_tags(name string, group string) string {
	return fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type = "WAF"

    managed_service_data {
      type = "WAF"

      rule_groups {
        id = "${aws_wafregional_rule_group.test.id}"

        override_action={
          type = "COUNT"
        }
      }

      default_action={
        type = "BLOCK"
      }
    }
  }
  resource_tags = {
    "Environment" = "Testing",
    "Usage"= "original",
  }

}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = %[2]q
}
`, name, group)
}

func testAccFmsPolicyConfig_tagsChanged(name string, group string) string {
	return fmt.Sprintf(`
resource "aws_fms_policy" "test" {
  exclude_resource_tags = false
  name                  = %[1]q
  remediation_enabled   = false
  resource_type_list    = ["AWS::ElasticLoadBalancingV2::LoadBalancer"]

  security_service_policy_data {
    type = "WAF"

    managed_service_data {
      type = "WAF"

      rule_groups {
        id = "${aws_wafregional_rule_group.test.id}"

        override_action={
          type = "COUNT"
        }
      }

      default_action={
        type = "BLOCK"
      }
    }
  }
  resource_tags = {
    "Usage"= "changed",
  }

}

resource "aws_wafregional_rule_group" "test" {
  metric_name = "MyTest"
  name        = %[2]q
}
`, name, group)
}
