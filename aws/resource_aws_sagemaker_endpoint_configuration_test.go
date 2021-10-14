package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_endpoint_configuration", &resource.Sweeper{
		Name: "aws_sagemaker_endpoint_configuration",
		Dependencies: []string{
			"aws_sagemaker_model",
		},
		F: testSweepSagemakerEndpointConfigurations,
	})
}

func testSweepSagemakerEndpointConfigurations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).sagemakerconn
	var sweeperErrs *multierror.Error

	req := &sagemaker.ListEndpointConfigsInput{
		NameContains: aws.String("tf-acc-test"),
	}
	err = conn.ListEndpointConfigsPages(req, func(page *sagemaker.ListEndpointConfigsOutput, lastPage bool) bool {
		for _, endpointConfig := range page.EndpointConfigs {

			r := resourceAwsSagemakerEndpointConfiguration()
			d := r.Data(nil)
			d.SetId(aws.StringValue(endpointConfig.EndpointConfigName))
			err := r.Delete(d, client)
			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Endpoint Config sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Sagemaker Endpoint Configs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSagemakerEndpointConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "production_variants.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.variant_name", "variant-1"),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.model_name", rName),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.initial_instance_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.instance_type", "ml.t2.medium"),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.initial_variant_weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.code_dump_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.#", "0"),
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

func TestAccAWSSagemakerEndpointConfiguration_productionVariants_InitialVariantWeight(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfig_ProductionVariants_InitialVariantWeight(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "production_variants.1.initial_variant_weight", "0.5"),
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

func TestAccAWSSagemakerEndpointConfiguration_productionVariants_AcceleratorType(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfig_ProductionVariant_AcceleratorType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "production_variants.0.accelerator_type", "ml.eia1.medium"),
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

func TestAccAWSSagemakerEndpointConfiguration_kmsKeyId(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfiguration_Config_KmsKeyId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", "aws_kms_key.test", "arn"),
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

func TestAccAWSSagemakerEndpointConfiguration_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSagemakerEndpointConfigurationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSagemakerEndpointConfigurationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerEndpointConfiguration_dataCaptureConfig(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationDataCaptureConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.enable_capture", "true"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.initial_sampling_percentage", "50"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.destination_s3_uri", fmt.Sprintf("s3://%s/", rName)),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.capture_options.0.capture_mode", "Input"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.capture_options.1.capture_mode", "Output"),
					resource.TestCheckResourceAttr(resourceName, "data_capture_config.0.capture_content_type_header.0.json_content_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "data_capture_config.0.capture_content_type_header.0.json_content_types.*", "application/json"),
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

func TestAccAWSSagemakerEndpointConfiguration_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsSagemakerEndpointConfiguration(), resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsSagemakerEndpointConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSagemakerEndpointConfiguration_async(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfigAsyncConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.client_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.output_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "async_inference_config.0.output_config.0.s3_output_path"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.output_config.0.notification_config.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "async_inference_config.0.output_config.0.kms_key_id", "aws_kms_key.test", "arn"),
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

func TestAccAWSSagemakerEndpointConfiguration_async_notifConfig(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfigAsyncNotifConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.client_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.output_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "async_inference_config.0.output_config.0.s3_output_path"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.output_config.0.notification_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "async_inference_config.0.output_config.0.notification_config.0.error_topic", "aws_sns_topic.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "async_inference_config.0.output_config.0.notification_config.0.success_topic", "aws_sns_topic.test", "arn"),
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

func TestAccAWSSagemakerEndpointConfiguration_async_client(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSagemakerEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSagemakerEndpointConfigurationConfigAsyncClientConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSagemakerEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.client_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.client_config.0.max_concurrent_invocations_per_instance", "1"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.output_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "async_inference_config.0.output_config.0.s3_output_path"),
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

func testAccCheckSagemakerEndpointConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_endpoint_configuration" {
			continue
		}

		_, err := finder.EndpointConfigByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SageMaker Endpoint Configuration %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSagemakerEndpointConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SageMaker endpoint config not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SageMaker endpoint config ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		_, err := finder.EndpointConfigByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccSagemakerEndpointConfigurationConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "kmeans"
}

resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfig_Basic(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfig_ProductionVariants_InitialVariantWeight(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
  }

  production_variants {
    variant_name           = "variant-2"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 0.5
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfig_ProductionVariant_AcceleratorType(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    accelerator_type       = "ml.eia1.medium"
    initial_variant_weight = 1
  }
}
`, rName)
}

func testAccSagemakerEndpointConfiguration_Config_KmsKeyId(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name        = %[1]q
  kms_key_arn = aws_kms_key.test.arn

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 10
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccSagemakerEndpointConfigurationConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccSagemakerEndpointConfigurationDataCaptureConfig(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  data_capture_config {
    enable_capture              = true
    initial_sampling_percentage = 50
    destination_s3_uri          = "s3://${aws_s3_bucket.test.bucket}/"

    capture_options {
      capture_mode = "Input"
    }

    capture_options {
      capture_mode = "Output"
    }

    capture_content_type_header {
      json_content_types = ["application/json"]
    }
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfigAsyncConfig(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  async_inference_config {
    output_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
      kms_key_id     = aws_kms_key.test.arn
    }
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfigAsyncNotifConfig(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  async_inference_config {
    output_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
      kms_key_id     = aws_kms_key.test.arn

      notification_config {
        error_topic   = aws_sns_topic.test.arn
        success_topic = aws_sns_topic.test.arn
      }
    }
  }
}
`, rName)
}

func testAccSagemakerEndpointConfigurationConfigAsyncClientConfig(rName string) string {
	return testAccSagemakerEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.test.name
    initial_instance_count = 2
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  async_inference_config {
    client_config {
      max_concurrent_invocations_per_instance = 1
    }

    output_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
      kms_key_id     = aws_kms_key.test.arn
    }
  }
}
`, rName)
}
