package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_endpoint_configuration", &resource.Sweeper{
		Name: "aws_sagemaker_endpoint_configuration",
		Dependencies: []string{
			"aws_sagemaker_model",
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
		return fmt.Errorf("error listing endpoint configs: %s", err)
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
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigExists("aws_sagemaker_endpoint_configuration.foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo", "name",
						fmt.Sprintf("terraform-testacc-sagemaker-endpoint-config-%s", rName)),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.0.variant_name",
						"variant-1"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.0.model_name",
						fmt.Sprintf("terraform-testacc-sagemaker-model-%s", rName)),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.0.initial_instance_count",
						"2"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.0.instance_type",
						"ml.t2.medium"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.0.initial_variant_weight",
						"1"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_endpoint_configuration.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerEndpointConfig_productionVariants_initialVariantWeight(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigProductionVariantInitialVariantWeightConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigExists("aws_sagemaker_endpoint_configuration.foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.1.initial_variant_weight",
						"0.5"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_endpoint_configuration.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerEndpointConfig_productionVariants_acceleratorType(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigProductionVariantAcceleratorTypeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigExists("aws_sagemaker_endpoint_configuration.foo"),
					resource.TestCheckResourceAttr(
						"aws_sagemaker_endpoint_configuration.foo",
						"production_variants.0.accelerator_type",
						"ml.eia1.medium"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_endpoint_configuration.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerEndpointConfig_kmsKeyId(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigKmsKeyIdConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigExists("aws_sagemaker_endpoint_configuration.foo"),
					resource.TestCheckResourceAttrSet("aws_sagemaker_endpoint_configuration.foo", "kms_key_id"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_endpoint_configuration.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerEndpointConfig_tags(t *testing.T) {
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigExists("aws_sagemaker_endpoint_configuration.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint_configuration.foo",
						"tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint_configuration.foo",
						"tags.foo", "bar"),
				),
			},
			{
				Config: testAccSagemakerEndpointConfigTagsUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigExists("aws_sagemaker_endpoint_configuration.foo"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint_configuration.foo",
						"tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sagemaker_endpoint_configuration.foo",
						"tags.bar", "baz"),
				),
			},
			{
				ResourceName:      "aws_sagemaker_endpoint_configuration.foo",
				ImportState:       true,
				ImportStateVerify: true,
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

func testAccCheckSagemakerEndpointConfigExists(n string) resource.TestCheckFunc {
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
		_, err := conn.DescribeEndpointConfig(opts)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccSagemakerEndpointConfigConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-%s"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 2
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"


	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
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
`, rName, rName, rName)
}

func testAccSagemakerEndpointConfigProductionVariantInitialVariantWeightConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-%s"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
	}

	production_variants {
		variant_name = "variant-2"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 0.5
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"


	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
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
`, rName, rName, rName)
}

func testAccSagemakerEndpointConfigProductionVariantAcceleratorTypeConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-%s"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 2
		instance_type = "ml.t2.medium"
		accelerator_type = "ml.eia1.medium"
		initial_variant_weight = 1
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"


	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
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
`, rName, rName, rName)
}

func testAccSagemakerEndpointConfigKmsKeyIdConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-%s"
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
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"


	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
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

resource "aws_kms_key" "foo" {
  description             = "terraform-testacc-sagemaker-model-%s"
  deletion_window_in_days = 10
}
`, rName, rName, rName, rName)
}

func testAccSagemakerEndpointConfigConfigTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-%s"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}

	tags = {
		foo = "bar"
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"


	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
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
`, rName, rName, rName)
}

func testAccSagemakerEndpointConfigTagsUpdateConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "foo" {
	name = "terraform-testacc-sagemaker-endpoint-config-%s"

	production_variants {
		variant_name = "variant-1"
		model_name = "${aws_sagemaker_model.foo.name}"
		initial_instance_count = 1
		instance_type = "ml.t2.medium"
		initial_variant_weight = 1
	}

	tags = {
		bar = "baz"
	}
}

resource "aws_sagemaker_model" "foo" {
	name = "terraform-testacc-sagemaker-model-%s"
	execution_role_arn = "${aws_iam_role.foo.arn}"

	primary_container {
		image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
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
`, rName, rName, rName)
}
