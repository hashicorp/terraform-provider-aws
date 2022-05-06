package waf_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
)

func TestAccWAFWebACL_basic(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`webacl/.+`)),
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

func TestAccWAFWebACL_changeNameForceNew(t *testing.T) {
	var webACL waf.WebACL
	rName1 := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	rName2 := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_Required(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
				),
			},
			{
				Config: testAccWebACLConfig_Required(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
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

func TestAccWAFWebACL_defaultAction(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_DefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
				),
			},
			{
				Config: testAccWebACLConfig_DefaultAction(rName, "BLOCK"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "BLOCK"),
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

func TestAccWAFWebACL_rules(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			// Test creating with rule
			{
				Config: testAccWebACLConfig_Rules_Single_Rule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
				),
			},
			// Test adding rule
			{
				Config: testAccWebACLConfig_Rules_Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
				),
			},
			// Test removing rule
			{
				Config: testAccWebACLConfig_Rules_Single_RuleGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
				),
			},
			// Test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWAFWebACL_logging(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			testAccPreCheckLoggingConfiguration(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_Logging(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.0.field_to_match.#", "2"),
				),
			},
			// Test resource import
			{
				Config:            testAccWebACLConfig_Logging(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test logging configuration update
			{
				Config: testAccWebACLConfig_LoggingUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.#", "0"),
				),
			},
			// Test logging configuration removal
			{
				Config: testAccWebACLConfig_LoggingRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccWAFWebACL_disappears(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					acctest.CheckResourceDisappears(acctest.Provider, tfwaf.ResourceWebACL(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFWebACL_tags(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
				),
			},
			{
				Config: testAccWebACLTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
				),
			},
			{
				Config: testAccWebACLTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
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

func testAccCheckWebACLDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_web_acl" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetWebACL(
			&waf.GetWebACLInput{
				WebACLId: aws.String(rs.Primary.ID),
			})

		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading WAF Web ACL (%s): %w", rs.Primary.ID, err)
		}

		if resp != nil && resp.WebACL != nil {
			return fmt.Errorf("WAF Web ACL (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckWebACLExists(n string, v *waf.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WebACL ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetWebACL(&waf.GetWebACLInput{
			WebACLId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.WebACL.WebACLId == rs.Primary.ID {
			*v = *resp.WebACL
			return nil
		}

		return fmt.Errorf("WebACL (%s) not found", rs.Primary.ID)
	}
}

func testAccWebACLConfig_Required(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }
}
`, rName)
}

func testAccWebACLConfig_DefaultAction(rName, defaultAction string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = %q
  }
}
`, rName, defaultAction)
}

func testAccWebACLConfig_Rules_Single_Rule(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "test" {
  metric_name = %[1]q
  name        = %[1]q

  predicates {
    data_id = aws_waf_ipset.test.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  rules {
    priority = 1
    rule_id  = aws_waf_rule.test.id

    action {
      type = "BLOCK"
    }
  }
}
`, rName)
}

func testAccWebACLConfig_Rules_Single_RuleGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule_group" "test" {
  metric_name = %[1]q
  name        = %[1]q
}

resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  rules {
    priority = 1
    rule_id  = aws_waf_rule_group.test.id
    type     = "GROUP"

    override_action {
      type = "NONE"
    }
  }
}
`, rName)
}

func testAccWebACLConfig_Rules_Multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "test" {
  metric_name = %[1]q
  name        = %[1]q

  predicates {
    data_id = aws_waf_ipset.test.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_rule_group" "test" {
  metric_name = %[1]q
  name        = %[1]q
}

resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  rules {
    priority = 1
    rule_id  = aws_waf_rule.test.id

    action {
      type = "BLOCK"
    }
  }

  rules {
    priority = 2
    rule_id  = aws_waf_rule_group.test.id
    type     = "GROUP"

    override_action {
      type = "NONE"
    }
  }
}
`, rName)
}

func testAccWebACLConfig_Logging(rName string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  logging_configuration {
    log_destination = aws_kinesis_firehose_delivery_stream.test.arn

    redacted_fields {
      field_to_match {
        type = "URI"
      }

      field_to_match {
        data = "referer"
        type = "HEADER"
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  # the name must begin with aws-waf-logs-
  name        = "aws-waf-logs-%[1]s"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName))
}

func testAccWebACLConfig_LoggingRemoved(rName string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }
}
`, rName))
}

func testAccWebACLConfig_LoggingUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigurationRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  logging_configuration {
    log_destination = aws_kinesis_firehose_delivery_stream.test.arn
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  # the name must begin with aws-waf-logs-
  name        = "aws-waf-logs-%[1]s"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName))
}

func testAccWebACLTags1Config(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  tags = {
    %q = %q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccWebACLTags2Config(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
