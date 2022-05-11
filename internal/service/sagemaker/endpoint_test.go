package sagemaker_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSageMakerEndpoint_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("endpoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccSageMakerEndpoint_endpointName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_sagemaker_endpoint.test"
	sagemakerEndpointConfigurationResourceName1 := "aws_sagemaker_endpoint_configuration.test"
	sagemakerEndpointConfigurationResourceName2 := "aws_sagemaker_endpoint_configuration.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName1, "name"),
				),
			},
			{
				Config: testAccEndpointConfigEndpointConfigNameUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName2, "name"),
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

func TestAccSageMakerEndpoint_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
				),
			},
			{
				Config: testAccEndpointConfigTagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.bar", "baz"),
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

func TestAccSageMakerEndpoint_deploymentConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDeploymentBasicConfig(rName, "ALL_AT_ONCE", 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.auto_rollback_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.termination_wait_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.type", "ALL_AT_ONCE"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.wait_interval_in_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.canary_size.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.linear_step_size.#", "0"),
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

func TestAccSageMakerEndpoint_deploymentConfig_full(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDeploymentFullConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.auto_rollback_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.auto_rollback_configuration.0.alarms.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.termination_wait_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.type", "LINEAR"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.wait_interval_in_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.canary_size.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.linear_step_size.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.linear_step_size.0.type", "INSTANCE_COUNT"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.linear_step_size.0.value", "1"),
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

func TestAccSageMakerEndpoint_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceEndpoint(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEndpointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_endpoint" {
			continue
		}

		_, err := tfsagemaker.FindEndpointByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SageMaker Endpoint (%s) still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckEndpointExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SageMaker Endpoint ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		_, err := tfsagemaker.FindEndpointByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccEndpointConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "access" {
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
      "ecr:BatchGetImage",
      "s3:GetObject",
    ]

    resources = ["*"]
  }
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.access.json
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "model.tar.gz"
  source = "test-fixtures/sagemaker-tensorflow-serving-test-model.tar.gz"
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-tensorflow-serving"
  image_tag       = "1.12-cpu"
}

resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image          = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    model_data_url = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    initial_instance_count = 2
    initial_variant_weight = 1
    instance_type          = "ml.t2.medium"
    model_name             = aws_sagemaker_model.test.name
    variant_name           = "variant-1"
  }
}
`, rName)
}

func testAccEndpointConfig(rName string) string {
	return testAccEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q
}
`, rName)
}

func testAccEndpointConfigEndpointConfigNameUpdate(rName string) string {
	return testAccEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test2" {
  name = "%[1]s2"

  production_variants {
    initial_instance_count = 1
    initial_variant_weight = 1
    instance_type          = "ml.t2.medium"
    model_name             = aws_sagemaker_model.test.name
    variant_name           = "variant-1"
  }
}

resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test2.name
  name                 = %[1]q
}
`, rName)
}

func testAccEndpointConfigTags(rName string) string {
	return testAccEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q

  tags = {
    foo = "bar"
  }
}
`, rName)
}

func testAccEndpointConfigTagsUpdate(rName string) string {
	return testAccEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q

  tags = {
    bar = "baz"
  }
}
`, rName)
}

func testAccEndpointDeploymentBasicConfig(rName, tType string, wait int) string {
	return testAccEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q

  deployment_config {
    blue_green_update_policy {
      traffic_routing_configuration {
        type                     = %[2]q
        wait_interval_in_seconds = %[3]d
      }
    }
  }
}
`, rName, tType, wait)
}

func testAccEndpointDeploymentFullConfig(rName string) string {
	return testAccEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}

resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q

  deployment_config {
    blue_green_update_policy {
      traffic_routing_configuration {
        type                     = "LINEAR"
        wait_interval_in_seconds = "60"

        linear_step_size {
          type  = "INSTANCE_COUNT"
          value = 1
        }
      }
    }

    auto_rollback_configuration {
      alarms {
        alarm_name = aws_cloudwatch_metric_alarm.test.alarm_name
      }
    }
  }
}
`, rName)
}
