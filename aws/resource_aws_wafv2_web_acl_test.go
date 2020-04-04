package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsWafv2WebACL_Basic(t *testing.T) {
	var v wafv2.WebACL
	webACLName := fmt.Sprintf("web-acl-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_Basic(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists("aws_wafv2_web_acl.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", webACLName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_BasicUpdate(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists("aws_wafv2_web_acl.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.name", "rule-2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.priority", "10"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.action.0.count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.0.size_constraint_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.0.size_constraint_statement.0.comparison_operator", "LT"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.0.size_constraint_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.0.size_constraint_statement.0.field_to_match.0.query_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.0.size_constraint_statement.0.size", "50"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.0.size_constraint_statement.0.text_transformation.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.0.size_constraint_statement.0.text_transformation.2212084700.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.0.size_constraint_statement.0.text_transformation.2212084700.type", "CMD_LINE"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.0.size_constraint_statement.0.text_transformation.4292479961.priority", "5"),
					resource.TestCheckResourceAttr(resourceName, "rule.1707625671.statement.0.size_constraint_statement.0.text_transformation.4292479961.type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "rule.479520631.name", "rule-1"),
					resource.TestCheckResourceAttr(resourceName, "rule.479520631.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.479520631.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.479520631.action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.479520631.action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.479520631.action.0.count.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_ChangeNameForceNew(t *testing.T) {
	var before, after wafv2.WebACL
	webACLName := fmt.Sprintf("web-acl-%s", acctest.RandString(5))
	ruleGroupNewName := fmt.Sprintf("web-acl-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_Basic(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists("aws_wafv2_web_acl.test", &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", webACLName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_Basic(ruleGroupNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists("aws_wafv2_web_acl.test", &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_Minimal(t *testing.T) {
	var v wafv2.WebACL
	webACLName := fmt.Sprintf("web-acl-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_Minimal(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists("aws_wafv2_web_acl.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_Tags(t *testing.T) {
	var v wafv2.WebACL
	webACLName := fmt.Sprintf("web-acl-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_OneTag(webACLName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists("aws_wafv2_web_acl.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_TwoTags(webACLName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists("aws_wafv2_web_acl.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_OneTag(webACLName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists("aws_wafv2_web_acl.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

func testAccCheckAwsWafv2WebACLDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_web_acl" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetWebACL(
			&wafv2.GetWebACLInput{
				Id:    aws.String(rs.Primary.ID),
				Name:  aws.String(rs.Primary.Attributes["name"]),
				Scope: aws.String(rs.Primary.Attributes["scope"]),
			})

		if err == nil {
			if *resp.WebACL.Id == rs.Primary.ID {
				return fmt.Errorf("WAFV2 WebACL %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the WebACL is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == wafv2.ErrCodeWAFNonexistentItemException {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckAwsWafv2WebACLExists(n string, v *wafv2.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFV2 WebACL ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetWebACL(&wafv2.GetWebACLInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err != nil {
			return err
		}

		if *resp.WebACL.Id == rs.Primary.ID {
			*v = *resp.WebACL
			return nil
		}

		return fmt.Errorf("WAFV2 WebACL (%s) not found", rs.Primary.ID)
	}
}

func testAccAwsWafv2WebACLConfig_Basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name = "%s"
  description = "%s"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name, name)
}

func testAccAwsWafv2WebACLConfig_BasicUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name = "%s"
  description = "Updated"
  scope = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name = "rule-2"
    priority = 10

    action {
      count {}
    }

    statement {
      size_constraint_statement {
        comparison_operator = "LT"
        size = 50

        field_to_match {
          query_string {}
        }

        text_transformation {
          priority = 5
          type = "NONE"
        }

        text_transformation {
          priority = 2
          type = "CMD_LINE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  rule {
    name = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_Minimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name = "%s"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_OneTag(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name = "%s"
  description = "%s"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }

  tags = {
    %q = %q
  }
}
`, name, name, tagKey, tagValue)
}

func testAccAwsWafv2WebACLConfig_TwoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name = "%s"
  description = "%s"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, name, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAwsWafv2WebACLImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
