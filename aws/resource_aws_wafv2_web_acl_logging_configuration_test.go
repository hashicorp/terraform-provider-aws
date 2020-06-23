package aws

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsWafv2WebACLLoggingConfiguration_basic(t *testing.T) {
	var v wafv2.LoggingConfiguration
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	rInt := acctest.RandInt()
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	kinesisResourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rInt, webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "resource_arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					// TODO: determine Set Hash # or use tfawsresource method if available
					resource.TestCheckResourceAttrPair(resourceName, "log_destination_configs.12345", kinesisResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
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

func TestAccAwsWafv2WebACLLoggingConfiguration_update(t *testing.T) {
	var v wafv2.LoggingConfiguration
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	rInt := acctest.RandInt()
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	kinesisResourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rInt, webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "resource_arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					// TODO: determine Set Hash # or use tfawsresource method if available
					resource.TestCheckResourceAttrPair(resourceName, "log_destination_configs.12345", kinesisResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_updateTwoRedactedFields(rInt, webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "resource_arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					// TODO: determine Set Hash # or use tfawsresource method if available
					resource.TestCheckResourceAttrPair(resourceName, "log_destination_configs.12345", kinesisResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"name":  "single_query_argument.0.name",
						"value": "username",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"name":  "single_header.0.name",
						"value": "user-agent",
					}),
				),
			},
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_updateOneRedactedField(rInt, webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "resource_arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					// TODO: determine Set Hash # or use tfawsresource method if available
					resource.TestCheckResourceAttrPair(resourceName, "log_destination_configs.12345", kinesisResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"name":  "single_header.0.name",
						"value": "user-agent",
					}),
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

func TestAccAwsWafv2WebACL_changeResourceArnForceNew(t *testing.T) {
	var before, after wafv2.LoggingConfiguration
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	webACLNameNew := acctest.RandomWithPrefix("tf-acc-test")
	rInt := acctest.RandInt()
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"
	kinesisResourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rInt, webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &before),
					resource.TestCheckResourceAttr(kinesisResourceName, "name", webACLName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					// TODO: determine Set Hash # or use tfawsresource method if available
					resource.TestCheckResourceAttrPair(resourceName, "log_destination_configs.12345", kinesisResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rInt, webACLNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &after),
					resource.TestCheckResourceAttr(webACLResourceName, "name", webACLNameNew),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					// TODO: determine Set Hash # or use tfawsresource method if available
					resource.TestCheckResourceAttrPair(resourceName, "log_destination_configs.12345", kinesisResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
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

func TestAccAwsWafv2WebACL_changeLogDestinationConfigsForceNew(t *testing.T) {
	var before, after wafv2.LoggingConfiguration
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	rInt := acctest.RandInt()
	rIntTwo := acctest.RandInt()
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"
	kinesisResourceName := "aws_kinesis_firehose_delivery_stream.test"
	kinesisResourceNameFoo := "aws_kinesis_firehose_delivery_stream.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rInt, webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &before),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					// TODO: determine Set Hash # or use tfawsresource method if available
					resource.TestCheckResourceAttrPair(resourceName, "log_destination_configs.12345", kinesisResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_multipleLoggingConfigs(rInt, rIntTwo, webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &after),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "2"),
					// TODO: determine Set Hash # or use tfawsresource method if available
					resource.TestCheckResourceAttrPair(resourceName, "log_destination_configs.12345", kinesisResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination_configs.45678", kinesisResourceNameFoo, "arn"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
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

func TestAccAwsWafv2WebACLLoggingConfiguration_disappears(t *testing.T) {
	var v wafv2.LoggingConfiguration
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	rInt := acctest.RandInt()
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rInt, webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafv2WebACLLoggingConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rInt, webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafv2WebACL(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_web_acl_logging_configuration" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetLoggingConfiguration(
			&wafv2.GetLoggingConfigurationInput{
				ResourceArn: aws.String(rs.Primary.ID),
			})

		if err != nil {
			// Return nil if the WebACL Logging Configuration is already destroyed
			if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
				return nil
			}
			return err
		}

		if resp == nil || resp.LoggingConfiguration == nil {
			return fmt.Errorf("Error getting WAFv2 WebACL Logging Configuration")
		}
		if aws.StringValue(resp.LoggingConfiguration.ResourceArn) == rs.Primary.ID {
			return fmt.Errorf("WAFv2 WebACL Logging Configuration for WebACL ARN %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsWafv2WebACLLoggingConfigurationExists(n string, v *wafv2.LoggingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 WebACL Logging Configuration ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetLoggingConfiguration(&wafv2.GetLoggingConfigurationInput{
			ResourceArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.LoggingConfiguration == nil {
			return fmt.Errorf("Error getting WAFv2 WebACL Logging Configuration")
		}

		if aws.StringValue(resp.LoggingConfiguration.ResourceArn) == rs.Primary.ID {
			*v = *resp.LoggingConfiguration
			return nil
		}

		return fmt.Errorf("WAFv2 WebACL Logging Configuration (%s) not found", rs.Primary.ID)
	}
}

func testAccKinesisFirehoseDeliveryStreamDependencyConfig(rInt int, name string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "firehose" {
  name = "tf_acctest_firehose_delivery_role_%[1]d"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "%[2]s" {
  bucket = "tf-test-bucket-%[1]d"
  acl = "private"
}

resource "aws_iam_role_policy" "%[2]s" {
  name = "tf_acctest_firehose_delivery_policy_%[1]d"
  role = "${aws_iam_role.firehose.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.%[2]s.arn}",
        "${aws_s3_bucket.%[2]s.arn}/*"
      ]
    }
  ]
}
EOF
}
`, rInt, name)
}

func testAccKinesisFirehoseDeliveryStreamConfig(rInt int, name string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "%[2]s" {
	depends_on = ["aws_iam_role_policy.firehose"]
	name = "aws-waf-logs-%[1]d"
	destination = "s3"
	s3_configuration {
		role_arn = "${aws_iam_role.firehose.arn}"
		bucket_arn = "${aws_s3_bucket.%[2]s.arn}"
	}
}
`, rInt, name)
}

func testAccWebACLResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

const testAccWebACLLoggingConfigurationResourceConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]
  redacted_fields {}
}
`

const testAccWebACLLoggingConfigurationMultipleLoggingConfigs = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn, aws_kinesis_firehose_delivery_stream.foo.arn]
  redacted_fields {}
}
`

const testAccWebACLLoggingConfigurationResourceUpdateTwoRedactedFieldsConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]
  
  redacted_fields {
 	single_header {
      name = "user-agent"
    }
  }
  
  redacted_fields {
    single_query_argument {
	  name = "username"
    }
  }
}
`

const testAccWebACLLoggingConfigurationResourceUpdateOneRedactedFieldConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  redacted_fields {
    single_header {
      name = "user-agent"
    }
  }
}
`

func testAccAwsWafv2WebACLLoggingConfiguration_basic(rInt int, webACLName string) string {
	return composeConfig(
		testAccKinesisFirehoseDeliveryStreamDependencyConfig(rInt, "test"),
		testAccKinesisFirehoseDeliveryStreamConfig(rInt, "test"),
		testAccWebACLResourceConfig(webACLName),
		testAccWebACLLoggingConfigurationResourceConfig)
}

func testAccAwsWafv2WebACLLoggingConfiguration_updateTwoRedactedFields(rInt int, webACLName string) string {
	return composeConfig(
		testAccKinesisFirehoseDeliveryStreamDependencyConfig(rInt, "test"),
		testAccKinesisFirehoseDeliveryStreamConfig(rInt, "test"),
		testAccWebACLResourceConfig(webACLName),
		testAccWebACLLoggingConfigurationResourceUpdateTwoRedactedFieldsConfig)
}

func testAccAwsWafv2WebACLLoggingConfiguration_updateOneRedactedField(rInt int, webACLName string) string {
	return composeConfig(
		testAccKinesisFirehoseDeliveryStreamDependencyConfig(rInt, "test"),
		testAccKinesisFirehoseDeliveryStreamConfig(rInt, "test"),
		testAccWebACLResourceConfig(webACLName),
		testAccWebACLLoggingConfigurationResourceUpdateOneRedactedFieldConfig)
}

func testAccAwsWafv2WebACLLoggingConfiguration_multipleLoggingConfigs(rInt, rIntTwo int, webACLName string) string {
	return composeConfig(
		testAccKinesisFirehoseDeliveryStreamDependencyConfig(rInt, "test"),
		testAccKinesisFirehoseDeliveryStreamConfig(rInt, "test"),
		testAccKinesisFirehoseDeliveryStreamDependencyConfig(rIntTwo, "foo"),
		testAccKinesisFirehoseDeliveryStreamConfig(rIntTwo, "foo"),
		testAccWebACLResourceConfig(webACLName),
		testAccWebACLLoggingConfigurationMultipleLoggingConfigs)
}
