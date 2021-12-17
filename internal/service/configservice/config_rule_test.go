package configservice_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccConfigRule_basic(t *testing.T) {
	var cr configservice.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &cr),
					testAccCheckConfigRuleName(resourceName, rName, &cr),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_identifier", "S3_BUCKET_VERSIONING_ENABLED"),
				),
			},
		},
	})
}

func testAccConfigRule_ownerAws(t *testing.T) {
	var cr configservice.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_ownerAws(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &cr),
					testAccCheckConfigRuleName(resourceName, rName, &cr),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile("config-rule/config-rule-[a-z0-9]+$")),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "rule_id", regexp.MustCompile("config-rule-[a-z0-9]+$")),
					resource.TestCheckResourceAttr(resourceName, "description", "Terraform Acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_identifier", "REQUIRED_TAGS"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_detail.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.compliance_resource_id", "blablah"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.compliance_resource_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.compliance_resource_types.*", "AWS::EC2::Instance"),
				),
			},
		},
	})
}

func testAccConfigRule_customlambda(t *testing.T) {
	var cr configservice.ConfigRule
	rInt := sdkacctest.RandInt()
	resourceName := "aws_config_config_rule.test"

	expectedName := fmt.Sprintf("tf-acc-test-%d", rInt)
	path := "test-fixtures/lambdatest.zip"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_customLambda(rInt, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &cr),
					testAccCheckConfigRuleName(resourceName, expectedName, &cr),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile("config-rule/config-rule-[a-z0-9]+$")),
					resource.TestCheckResourceAttr(resourceName, "name", expectedName),
					resource.TestMatchResourceAttr(resourceName, "rule_id", regexp.MustCompile("config-rule-[a-z0-9]+$")),
					resource.TestCheckResourceAttr(resourceName, "description", "Terraform Acceptance tests"),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "Six_Hours"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.owner", "CUSTOM_LAMBDA"),
					resource.TestCheckResourceAttrPair(resourceName, "source.0.source_identifier", "aws_lambda_function.f", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_detail.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "source.0.source_detail.*", map[string]string{
						"event_source":                "aws.config",
						"message_type":                "ConfigurationSnapshotDeliveryCompleted",
						"maximum_execution_frequency": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_key", "IsTemporary"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_value", "yes"),
				),
			},
		},
	})
}

func testAccConfigRule_importAws(t *testing.T) {
	resourceName := "aws_config_config_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_ownerAws(rName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccConfigRule_importLambda(t *testing.T) {
	resourceName := "aws_config_config_rule.test"
	rInt := sdkacctest.RandInt()

	path := "test-fixtures/lambdatest.zip"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_customLambda(rInt, path),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccConfigRule_Scope_TagKey(t *testing.T) {
	var configRule configservice.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_Scope_TagKey(rName, "key1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &configRule),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_key", "key1"),
				),
			},
			{
				Config: testAccConfigRuleConfig_Scope_TagKey(rName, "key2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &configRule),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_key", "key2"),
				),
			},
		},
	})
}

func testAccConfigRule_Scope_TagKey_Empty(t *testing.T) {
	var configRule configservice.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_Scope_TagKey(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &configRule),
				),
			},
		},
	})
}

func testAccConfigRule_Scope_TagValue(t *testing.T) {
	var configRule configservice.ConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_config_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_Scope_TagValue(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &configRule),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_value", "value1"),
				),
			},
			{
				Config: testAccConfigRuleConfig_Scope_TagValue(rName, "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &configRule),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.tag_value", "value2"),
				),
			},
		},
	})
}

func testAccConfigRule_tags(t *testing.T) {
	var cr configservice.ConfigRule
	resourceName := "aws_config_config_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigRuleConfig_Tags(rName, "foo", "bar", "fizz", "buzz"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccConfigRuleConfig_Tags(rName, "foo", "bar2", "fizz2", "buzz2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz2", "buzz2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigRuleExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckConfigRuleName(n, desired string, obj *configservice.ConfigRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.Attributes["name"] != *obj.ConfigRuleName {
			return fmt.Errorf("Expected name: %q, given: %q", desired, *obj.ConfigRuleName)
		}
		return nil
	}
}

func testAccCheckConfigRuleExists(n string, obj *configservice.ConfigRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No config rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn
		out, err := conn.DescribeConfigRules(&configservice.DescribeConfigRulesInput{
			ConfigRuleNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})
		if err != nil {
			return fmt.Errorf("Failed to describe config rule: %s", err)
		}
		if len(out.ConfigRules) < 1 {
			return fmt.Errorf("No config rule found when describing %q", rs.Primary.Attributes["name"])
		}

		cr := out.ConfigRules[0]
		*obj = *cr

		return nil
	}
}

func testAccCheckConfigRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_config_rule" {
			continue
		}

		resp, err := conn.DescribeConfigRules(&configservice.DescribeConfigRulesInput{
			ConfigRuleNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if err == nil {
			if len(resp.ConfigRules) != 0 &&
				*resp.ConfigRules[0].ConfigRuleName == rs.Primary.Attributes["name"] {
				return fmt.Errorf("config rule still exists: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccConfigRuleConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_config_configuration_recorder" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
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
        "Service": "config.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRole"
  role       = aws_iam_role.test.name
}
`, rName)
}

func testAccConfigRuleConfig_basic(rName string) string {
	return testAccConfigRuleConfig_base(rName) + fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %q

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName)
}

func testAccConfigRuleConfig_ownerAws(rName string) string {
	return testAccConfigRuleConfig_base(rName) + fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name        = %q
  description = "Terraform Acceptance tests"

  source {
    owner             = "AWS"
    source_identifier = "REQUIRED_TAGS"
  }

  scope {
    compliance_resource_id    = "blablah"
    compliance_resource_types = ["AWS::EC2::Instance"]
  }

  input_parameters = <<PARAMS
{
  "tag1Key": "CostCenter",
  "tag2Key": "Owner"
}
PARAMS

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName)
}

func testAccConfigRuleConfig_customLambda(randInt int, path string) string {
	return fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name                        = "tf-acc-test-%[1]d"
  description                 = "Terraform Acceptance tests"
  maximum_execution_frequency = "Six_Hours"

  source {
    owner             = "CUSTOM_LAMBDA"
    source_identifier = aws_lambda_function.f.arn

    source_detail {
      event_source = "aws.config"
      message_type = "ConfigurationSnapshotDeliveryCompleted"
    }
  }

  scope {
    tag_key   = "IsTemporary"
    tag_value = "yes"
  }

  depends_on = [
    aws_config_configuration_recorder.foo,
    aws_config_delivery_channel.foo,
  ]
}

resource "aws_lambda_function" "f" {
  filename      = "%[2]s"
  function_name = "tf_acc_lambda_awsconfig_%[1]d"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

data "aws_partition" "current" {}

resource "aws_lambda_permission" "p" {
  statement_id  = "AllowExecutionFromConfig"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.f.arn
  principal     = "config.${data.aws_partition.current.dns_suffix}"
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "tf_acc_lambda_aws_config_%[1]d"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "a" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRulesExecutionRole"
}

resource "aws_config_delivery_channel" "foo" {
  name           = "tf-acc-test-%[1]d"
  s3_bucket_name = aws_s3_bucket.b.bucket

  snapshot_delivery_properties {
    delivery_frequency = "Six_Hours"
  }

  depends_on = [aws_config_configuration_recorder.foo]
}

resource "aws_s3_bucket" "b" {
  bucket        = "tf-acc-awsconfig-%[1]d"
  force_destroy = true
}

resource "aws_config_configuration_recorder" "foo" {
  name     = "tf-acc-test-%[1]d"
  role_arn = aws_iam_role.r.arn
}

resource "aws_iam_role" "r" {
  name = "tf-acc-test-awsconfig-%[1]d"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "p" {
  name = "tf-acc-test-awsconfig-%[1]d"
  role = aws_iam_role.r.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "config:Put*",
      "Effect": "Allow",
      "Resource": "*"
    },
    {
      "Action": "s3:*",
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.b.arn}",
        "${aws_s3_bucket.b.arn}/*"
      ]
    },
    {
      "Action": "lambda:*",
      "Effect": "Allow",
      "Resource": "${aws_lambda_function.f.arn}"
    }
  ]
}
POLICY
}
`, randInt, path)
}

func testAccConfigRuleConfig_Scope_TagKey(rName, tagKey string) string {
	return testAccConfigRuleConfig_base(rName) + fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %q

  scope {
    tag_key = %q
  }

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName, tagKey)
}

func testAccConfigRuleConfig_Scope_TagValue(rName, tagValue string) string {
	return testAccConfigRuleConfig_base(rName) + fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %q

  scope {
    tag_key   = "key"
    tag_value = %q
  }

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName, tagValue)
}

func testAccConfigRuleConfig_Tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccConfigRuleConfig_base(rName) + fmt.Sprintf(`
resource "aws_config_config_rule" "test" {
  name = %[1]q

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  tags = {
    Name = %[1]q

    %[2]s = %[3]q
    %[4]s = %[5]q
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
