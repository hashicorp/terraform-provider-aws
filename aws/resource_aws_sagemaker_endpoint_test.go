package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists("aws_sagemaker_endpoint.foo"),

					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"name",
						fmt.Sprintf("terraform-testacc-sagemaker-endpoint-%s", rName)),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"endpoint_config_name",
						fmt.Sprintf("terraform-testacc-sagemaker-endpoint-config-%s", rName)),
				),
				ExpectError: regexp.MustCompile(`.*unexpected state 'Failed', wanted target 'InService'.*`),
			},
			// Does not work with failing resource
			//{
			//	ResourceName:      "aws_sagemaker_endpoint.foo",
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//},
		},
	})
}

func TestAccAWSSagemakerEndpoint_tags(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists("aws_sagemaker_endpoint.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint.foo", "tags.foo", "bar"),
				),
				ExpectError: regexp.MustCompile(`.*unexpected state 'Failed', wanted target 'InService'.*`),
			},
			{
				Config: testAccSagemakerEndpointConfigTagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists("aws_sagemaker_endpoint.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint.foo", "tags.bar", "baz"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_endpoint.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerEndpoint_update(t *testing.T) {
	// Cannot update failed endpoint
	t.Skip()

	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists("aws_sagemaker_endpoint.foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"name",
						fmt.Sprintf("terraform-testacc-sagemaker-endpoint-%s", rName)),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"endpoint_config_name",
						"terraform-testacc-sagemaker-endpoint-config-foo"),
				),
				ExpectError: regexp.MustCompile(`.*unexpected state 'Failed', wanted target 'InService'.*`),
			},
			{
				Config: testAccSagemakerEndpointConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointExists("aws_sagemaker_endpoint.foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"name",
						"terraform-testacc-sagemaker-endpoint-foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint.foo",
						"endpoint_config_name",
						fmt.Sprintf("terraform-testacc-sagemaker-endpoint-updated-%s", rName)),
				),
			},
			{
				ResourceName:      "aws_sagemaker_endpoint.foo",
				ImportState:       true,
				ImportStateVerify: true,
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

func testAccCheckSagemakerEndpointExists(n string) resource.TestCheckFunc {
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
		_, err := conn.DescribeEndpoint(opts)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccSagemakerEndpointConfig(rName string) string {
	return fmt.Sprintf(`resource "aws_sagemaker_endpoint" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-%s"
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.foo.name}"
}

resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-%s"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "%s"
	}
}

resource "aws_iam_role" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
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
  name = "terraform-testacc-sagemaker-endpoint-%s"
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
`, rName, rName, rName, image, rName, rName)
}

func testAccSagemakerEndpointConfigUpdate(rName string) string {
	return fmt.Sprintf(`resource "aws_sagemaker_endpoint" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-%s"
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.foo_updated.name}"
}

resource "aws_sagemaker_endpoint_configuration" "foo_updated" {
	name = "terraform-testacc-sagemaker-endpoint-config-updated-%s"

	production_variants {
		variant_name = "variant-2"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 2
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}
}

resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-%s"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "%s"
	}
}

resource "aws_iam_role" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
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
`, rName, rName, rName, rName, image, rName)
}

func testAccSagemakerEndpointConfigTags(rName string) string {
	return fmt.Sprintf(`resource "aws_sagemaker_endpoint" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-%s"
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.foo.name}"

	tags {
		foo = "bar"
	}
}

resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-%s"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "%s"
	}
}

resource "aws_iam_role" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
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
`, rName, rName, rName, image, rName)
}

func testAccSagemakerEndpointConfigTagsUpdate(rName string) string {
	return fmt.Sprintf(`resource "aws_sagemaker_endpoint" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-%s"
	endpoint_config_name = "${aws_sagemaker_endpoint_configuration.foo.name}"

	tags {
		bar = "baz"
	}
}

resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-%s"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "%s"
	}
}

resource "aws_iam_role" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
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
`, rName, rName, rName, image, rName)
}
