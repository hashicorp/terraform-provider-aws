package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"regexp"
	"testing"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_endpoint", &resource.Sweeper{
		Name: "aws_sagemaker_endpoint",
		Dependencies: []string{
			"aws_sagemaker_model",
			"aws_sagemaker_endpoint_configuration",
		},
		F: testSweepSagemakerEndpoints,
	})
}

func testSweepSagemakerEndpoints(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	req := &sagemaker.ListEndpointsInput{
		NameContains: aws.String("terraform-testacc-sagemaker-endpoint"),
	}
	resp, err := conn.ListEndpoints(req)
	if err != nil {
		return fmt.Errorf("error listing endpoints: %s", err)
	}

	if len(resp.Endpoints) == 0 {
		log.Print("[DEBUG] No SageMaker endpoint to sweep")
		return nil
	}

	for _, endpoint := range resp.Endpoints {
		_, err := conn.DeleteEndpoint(&sagemaker.DeleteEndpointInput{
			EndpointName: endpoint.EndpointName,
		})
		if err != nil {
			return fmt.Errorf(
				"error deleting sagemaker endpoint (%s): %s",
				*endpoint.EndpointName, err)
		}
	}

	return nil
}

func TestAccAWSSagemakerEndpoint_basic(t *testing.T) {
	var endpoint sagemaker.DescribeEndpointOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists("aws_sagemaker_endpoint.foo", &endpoint),
					testAccCheckSagemakerEndpointName(&endpoint, "terraform-testacc-sagemaker-endpoint-foo"),
					testAccCheckSagemakerEndpointEndpointConfigName(&endpoint,
						"terraform-testacc-sagemaker-endpoint-config-foo"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"name",
						"terraform-testacc-sagemaker-endpoint-foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"endpoint_config_name",
						"terraform-testacc-sagemaker-endpoint-config-foo"),
				),
				ExpectError: regexp.MustCompile(`.*unexpected state 'Failed', wanted target 'InService'.*`),
			},
		},
	})
}

func TestAccAWSSagemakerEndpoint_tags(t *testing.T) {
	var endpoint sagemaker.DescribeEndpointOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists("aws_sagemaker_endpoint.foo", &endpoint),
					testAccCheckSagemakerEndpointTags(&endpoint, "foo", "bar"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"name",
						"terraform-testacc-sagemaker-endpoint-foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint.foo", "tags.foo", "bar"),
				),
				ExpectError: regexp.MustCompile(`.*unexpected state 'Failed', wanted target 'InService'.*`),
			},

			{
				Config: testAccSagemakerEndpointConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists("aws_sagemaker_endpoint.foo", &endpoint),
					testAccCheckSagemakerEndpointTags(&endpoint, "foo", ""),
					testAccCheckSagemakerEndpointTags(&endpoint, "bar", "baz"),

					resource.TestCheckResourceAttr("aws_sagemaker_endpoint.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint.foo", "tags.bar", "baz"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerEndpoint_update(t *testing.T) {
	// Cannot update failed endpoint
	t.Skip()

	var endpoint sagemaker.DescribeEndpointOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists("aws_sagemaker_endpoint.foo", &endpoint),
					testAccCheckSagemakerEndpointName(&endpoint, "terraform-testacc-sagemaker-endpoint-foo"),
					testAccCheckSagemakerEndpointEndpointConfigName(&endpoint,
						"terraform-testacc-sagemaker-endpoint-config-foo"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"name",
						"terraform-testacc-sagemaker-endpoint-foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"endpoint_config_name",
						"terraform-testacc-sagemaker-endpoint-config-foo"),
				),
				ExpectError: regexp.MustCompile(`.*unexpected state 'Failed', wanted target 'InService'.*`),
			},
			{
				Config: testAccSagemakerEndpointConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists("aws_sagemaker_endpoint.foo", &endpoint),
					testAccCheckSagemakerEndpointName(&endpoint, "terraform-testacc-sagemaker-endpoint-foo"),
					testAccCheckSagemakerEndpointEndpointConfigName(&endpoint,
						"terraform-testacc-sagemaker-endpoint-config-foo-updated"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"name",
						"terraform-testacc-sagemaker-endpoint-foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"endpoint_config_name",
						"terraform-testacc-sagemaker-endpoint-config-foo-updated"),
				),
			},
		},
	})
}

func testAccCheckSagemakerEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_endpoint" {
			continue
		}

		resp, err := conn.ListEndpoints(&sagemaker.ListEndpointsInput{
			NameContains: aws.String(rs.Primary.ID),
		})
		if err == nil {
			if len(resp.Endpoints) > 0 {
				return fmt.Errorf("SageMaker endpoint still exists")
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

func testAccCheckSagemakerEndpointExists(n string, endpoint *sagemaker.DescribeEndpointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SageMaker endpoint ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		opts := &sagemaker.DescribeEndpointInput{
			EndpointName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeEndpoint(opts)
		if err != nil {
			return err
		}

		*endpoint = *resp

		return nil
	}
}

func testAccCheckSagemakerEndpointName(endpoint *sagemaker.DescribeEndpointOutput, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		endpointName := endpoint.EndpointName
		if *endpointName != expected {
			return fmt.Errorf("bad SageMaker endpoint name: %s", *endpoint.EndpointName)
		}

		return nil
	}
}

func testAccCheckSagemakerEndpointEndpointConfigName(endpoint *sagemaker.DescribeEndpointOutput, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := endpoint.EndpointConfigName
		if *name != expected {
			return fmt.Errorf("bad SageMaker endpoint config name: %s", *endpoint.EndpointConfigName)
		}

		return nil
	}
}

func testAccCheckSagemakerEndpointTags(endpoint *sagemaker.DescribeEndpointOutput, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

		ts, err := conn.ListTags(&sagemaker.ListTagsInput{
			ResourceArn: endpoint.EndpointArn,
		})
		if err != nil {
			return fmt.Errorf("failed listing tags: %s", err)
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

const testAccSagemakerEndpointConfig = `
resource "aws_sagemaker_endpoint" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-foo"
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.foo.name}"
}

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

resource "aws_iam_policy" "foo" {
  name = "terraform-testacc-sagemaker-endpoint-foo"
  description = "Allow SageMaker to create endpoint"
  policy = "${data.aws_iam_policy_document.foo.json}"
}

data "aws_iam_policy_document" "foo" {
  statement {
    effect = "Allow"
    actions = [
      "sagemaker:*"
    ]
    resources = [
      "*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:CreateLogGroup",
      "logs:DescribeLogStreams",
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage"
    ]
    resources = [
      "*"]
  }
}

resource "aws_iam_role_policy_attachment" "foo" {
  role = "${aws_iam_role.foo.name}"
  policy_arn = "${aws_iam_policy.foo.arn}"
}
`

const testAccSagemakerEndpointConfigUpdate = `
resource "aws_sagemaker_endpoint" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-foo"
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.foo_updated.name}"
}

resource "aws_sagemaker_endpoint_configuration" "foo_updated" {
	name = "terraform-testacc-sagemaker-endpoint-config-foo-updated"

	production_variants {
		variant_name = "variant-2"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 2
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}
}

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

const testAccSagemakerEndpointConfigTags = `
resource "aws_sagemaker_endpoint" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-foo"
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.foo.name}"

	tags {
		foo = "bar"
	}
}

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

const testAccSagemakerEndpointConfigTagsUpdate = `
resource "aws_sagemaker_endpoint" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-foo"
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.foo.name}"

	tags {
		bar = "baz"
	}
}

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
