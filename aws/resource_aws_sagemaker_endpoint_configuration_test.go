package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"testing"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_endpoint_configuration", &resource.Sweeper{
		Name: "aws_sagemaker_endpoint_configuration",
		Dependencies: []string{
			"aws_sagemaker_model",
			"aws_iam_role",
		},
		F: testSweepSagemakerEndpointConfigs,
	})
}

func testSweepSagemakerEndpointConfigs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	req := &sagemaker.ListEndpointConfigsInput{
		NameContains: aws.String("terraform-testacc-sagemaker-endpoint-config"),
	}
	resp, err := conn.ListEndpointConfigs(req)
	if err != nil {
		return fmt.Errorf("Error listing endpoint configs: %s", err)
	}

	if len(resp.EndpointConfigs) == 0 {
		log.Print("[DEBUG] No SageMaker endpoint config to sweep")
		return nil
	}

	for _, endpointConfig := range resp.EndpointConfigs {
		_, err := conn.DeleteEndpointConfig(&sagemaker.DeleteEndpointConfigInput{
			EndpointConfigName: endpointConfig.EndpointConfigName,
		})
		if err != nil {
			return fmt.Errorf(
				"failed to delete SageMaker endpoint config (%s): %s",
				*endpointConfig.EndpointConfigName, err)
		}
	}

	return nil
}

func TestAccAWSSagemakerEndpointConfig_basic(t *testing.T) {
	var endpointConfig sagemaker.DescribeEndpointConfigOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigExists("aws_sagemaker_endpoint_configuration.foo",
						&endpointConfig),
					testAccCheckSagemakerEndpointConfigName(&endpointConfig,
						"terraform-testacc-sagemaker-endpoint-config-foo"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo", "name",
						"terraform-testacc-sagemaker-endpoint-config-foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.2891507008.variant_name",
						"variant-1"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.2891507008.model_name",
						"terraform-testacc-sagemaker-model-foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.2891507008.initial_instance_count",
						"1"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.2891507008.instance_type",
						"ml.t2.medium"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.2891507008.initial_variant_weight",
						"1"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerEndpointConfig_kmsKeyId(t *testing.T) {
	var endpointConfig sagemaker.DescribeEndpointConfigOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigKmsKeyIdConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigExists("aws_sagemaker_endpoint_configuration.foo",
						&endpointConfig),
					testAccCheckSagemakerEndpointConfigKmsKeyId(&endpointConfig),

					resource.TestCheckResourceAttrSet("aws_sagemaker_endpoint_configuration.foo", "kms_key_id"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerEndpointConfig_tags(t *testing.T) {
	var endpointConfig sagemaker.DescribeEndpointConfigOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigExists("aws_sagemaker_endpoint_configuration.foo",
						&endpointConfig),
					testAccCheckSagemakerEndpointConfigTags(&endpointConfig, "foo", "bar"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo", "name",
						"terraform-testacc-sagemaker-endpoint-config-foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint_configuration.foo",
						"tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint_configuration.foo",
						"tags.foo", "bar"),
				),
			},

			{
				Config: testAccSagemakerEndpointConfigConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigExists("aws_sagemaker_endpoint_configuration.foo",
						&endpointConfig),
					testAccCheckSagemakerEndpointConfigTags(&endpointConfig, "foo", ""),
					testAccCheckSagemakerEndpointConfigTags(&endpointConfig, "bar", "baz"),

					resource.TestCheckResourceAttr("aws_sagemaker_endpoint_configuration.foo",
						"tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint_configuration.foo",
						"tags.bar", "baz"),
				),
			},
		},
	})
}

func testAccCheckSagemakerEndpointConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_endpoint_configuration" {
			continue
		}

		resp, err := conn.ListEndpointConfigs(&sagemaker.ListEndpointConfigsInput{
			NameContains: aws.String(rs.Primary.ID),
		})
		if err == nil {
			if len(resp.EndpointConfigs) > 0 {
				return fmt.Errorf("SageMaker endpoint configs still exists")
			}

			return nil
		}

		sagemakerErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if sagemakerErr.Code() != "ResourceNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckSagemakerEndpointConfigExists(n string,
	endpointConfig *sagemaker.DescribeEndpointConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SageMaker endpoint config not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SageMaker endpoint config ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		opts := &sagemaker.DescribeEndpointConfigInput{
			EndpointConfigName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeEndpointConfig(opts)
		if err != nil {
			return err
		}

		*endpointConfig = *resp
		return nil
	}
}

func testAccCheckSagemakerEndpointConfigName(endpointConfig *sagemaker.DescribeEndpointConfigOutput,
	expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := endpointConfig.EndpointConfigName
		if *name != expected {
			return fmt.Errorf("bad name: %s", *name)
		}

		return nil
	}
}

func testAccCheckSagemakerEndpointConfigKmsKeyId(endpointConfig *sagemaker.DescribeEndpointConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := endpointConfig.KmsKeyId
		if id == nil || len(*id) < 20 {
			return fmt.Errorf("bad KMS key ID: %s", *id)
		}

		return nil
	}
}

func testAccCheckSagemakerEndpointConfigTags(endpointConfig *sagemaker.DescribeEndpointConfigOutput,
	key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

		ts, err := conn.ListTags(&sagemaker.ListTagsInput{
			ResourceArn: endpointConfig.EndpointConfigArn,
		})
		if err != nil {
			return fmt.Errorf("failed to list tags: %s", err)
		}

		m := tagsToMapSagemaker(ts.Tags)
		v, ok := m[key]
		if value != "" && !ok {
			return fmt.Errorf("missing tag: %s", key)
		} else if value == "" && ok {
			return fmt.Errorf("extra tag: %s", key)
		}
		if value == "" {
			return nil
		}

		if v != value {
			return fmt.Errorf("%s: bad value: %s", key, v)
		}

		return nil
	}
}

const testAccSagemakerEndpointConfigConfig = `
resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-foo"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-foo"
	execution_role_arn = "${aws_iam_role.foo.arn}"


	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
	}
}

resource "aws_iam_role" "foo" {
  name = "terraform-testacc-sagemaker-model-foo"
  path = "/"
  assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = [ "sts:AssumeRole" ]
    principals {
      type = "Service"
      identifiers = [ "sagemaker.amazonaws.com" ]
    }
  }
}
`

const testAccSagemakerEndpointConfigKmsKeyIdConfig = `
resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-foo"
	kms_key_id = "${aws_kms_key.foo.arn}"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-foo"
	execution_role_arn = "${aws_iam_role.foo.arn}"


	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
	}
}

resource "aws_iam_role" "foo" {
  name = "terraform-testacc-sagemaker-model-foo"
  path = "/"
  assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = [ "sts:AssumeRole" ]
    principals {
      type = "Service"
      identifiers = [ "sagemaker.amazonaws.com" ]
    }
  }
}

resource "aws_kms_key" "foo" {
  description             = "terraform-testacc-sagemaker-model-foo"
  deletion_window_in_days = 10
}
`

const testAccSagemakerEndpointConfigConfigTags = `
resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-foo"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}

	tags {
		foo = "bar"
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-foo"
	execution_role_arn = "${aws_iam_role.foo.arn}"


	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
	}
}

resource "aws_iam_role" "foo" {
  name = "terraform-testacc-sagemaker-model-foo"
  path = "/"
  assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = [ "sts:AssumeRole" ]
    principals {
      type = "Service"
      identifiers = [ "sagemaker.amazonaws.com" ]
    }
  }
}
`

const testAccSagemakerEndpointConfigConfigTagsUpdate = `
resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-foo"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}

	tags {
		bar = "baz"
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-foo"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
	}
}

resource "aws_iam_role" "foo" {
  name = "terraform-testacc-sagemaker-model-foo"
  path = "/"
  assume_role_policy = "${data.aws_iam_policy_document.assume_role.json}"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = [ "sts:AssumeRole" ]
    principals {
      type = "Service"
      identifiers = [ "sagemaker.amazonaws.com" ]
    }
  }
}
`
