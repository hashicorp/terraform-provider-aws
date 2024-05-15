// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFWebACL_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL awstypes.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rules.#", acctest.Ct0),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "waf", regexache.MustCompile(`webacl/.+`)),
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
	ctx := acctest.Context(t)
	var webACL awstypes.WebACL
	rName1 := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	rName2 := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_required(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, rName1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rules.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLConfig_required(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rules.#", acctest.Ct0),
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
	ctx := acctest.Context(t)
	var webACL awstypes.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_defaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
				),
			},
			{
				Config: testAccWebACLConfig_defaultAction(rName, "BLOCK"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var webACL awstypes.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating with rule
			{
				Config: testAccWebACLConfig_rulesSingleRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", acctest.Ct1),
				),
			},
			// Test adding rule
			{
				Config: testAccWebACLConfig_rulesMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", acctest.Ct2),
				),
			},
			// Test removing rule
			{
				Config: testAccWebACLConfig_rulesSingleRuleGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var webACL awstypes.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_logging(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.0.field_to_match.#", acctest.Ct2),
				),
			},
			// Test resource import
			{
				Config:            testAccWebACLConfig_logging(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test logging configuration update
			{
				Config: testAccWebACLConfig_loggingUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.#", acctest.Ct0),
				),
			},
			// Test logging configuration removal
			{
				Config: testAccWebACLConfig_loggingRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccWAFWebACL_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL awstypes.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwaf.ResourceWebACL(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFWebACL_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL awstypes.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "rules.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, "rules.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, "rules.#", acctest.Ct0),
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

func testAccCheckWebACLDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_waf_web_acl" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

			_, err := tfwaf.FindWebACLByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Web ACL %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWebACLExists(ctx context.Context, n string, v *awstypes.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

		output, err := tfwaf.FindWebACLByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccWebACLConfig_required(rName string) string {
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

func testAccWebACLConfig_defaultAction(rName, defaultAction string) string {
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

func testAccWebACLConfig_rulesSingleRule(rName string) string {
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
  lifecycle {
    create_before_destroy = true
  }
}
`, rName)
}

func testAccWebACLConfig_rulesSingleRuleGroup(rName string) string {
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
  lifecycle {
    create_before_destroy = true
  }
}
`, rName)
}

func testAccWebACLConfig_rulesMultiple(rName string) string {
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
  lifecycle {
    create_before_destroy = true
  }
}
`, rName)
}

func testAccWebACLConfig_logging(rName string) string {
	return fmt.Sprintf(`
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
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName)
}

func testAccWebACLConfig_loggingRemoved(rName string) string {
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

func testAccWebACLConfig_loggingUpdate(rName string) string {
	return fmt.Sprintf(`
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
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName)
}

func testAccWebACLConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccWebACLConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
