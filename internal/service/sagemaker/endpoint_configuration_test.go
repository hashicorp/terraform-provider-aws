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

func TestAccSageMakerEndpointConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
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

func TestAccSageMakerEndpointConfiguration_ProductionVariants_initialVariantWeight(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationConfig_ProductionVariants_InitialVariantWeight(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
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

func TestAccSageMakerEndpointConfiguration_ProductionVariants_acceleratorType(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationConfig_ProductionVariant_AcceleratorType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
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

func TestAccSageMakerEndpointConfiguration_kmsKeyID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationConfig_kmsKeyId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
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

func TestAccSageMakerEndpointConfiguration_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
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
				Config: testAccEndpointConfigurationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEndpointConfigurationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerEndpointConfiguration_dataCapture(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationDataCaptureConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
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

func TestAccSageMakerEndpointConfiguration_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceEndpointConfiguration(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceEndpointConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSageMakerEndpointConfiguration_async(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationConfigAsyncConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.client_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.output_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "async_inference_config.0.output_config.0.s3_output_path"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.output_config.0.notification_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "async_inference_config.0.output_config.0.kms_key_id", ""),
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

func TestAccSageMakerEndpointConfiguration_async_kms(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationConfigAsyncKMSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
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

func TestAccSageMakerEndpointConfiguration_Async_notif(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationConfigAsyncNotifConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
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

func TestAccSageMakerEndpointConfiguration_Async_client(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEndpointConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfigurationConfigAsyncClientConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointConfigurationExists(resourceName),
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

func testAccCheckEndpointConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_endpoint_configuration" {
			continue
		}

		_, err := tfsagemaker.FindEndpointConfigByName(conn, rs.Primary.ID)

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

func testAccCheckEndpointConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SageMaker endpoint config not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SageMaker endpoint config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		_, err := tfsagemaker.FindEndpointConfigByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccEndpointConfigurationConfig_Base(rName string) string {
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

func testAccEndpointConfigurationConfig_Basic(rName string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
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

func testAccEndpointConfigurationConfig_ProductionVariants_InitialVariantWeight(rName string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
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

func testAccEndpointConfigurationConfig_ProductionVariant_AcceleratorType(rName string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
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

func testAccEndpointConfigurationConfig_kmsKeyId(rName string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
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

func testAccEndpointConfigurationConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
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

func testAccEndpointConfigurationConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
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

func testAccEndpointConfigurationDataCaptureConfig(rName string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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

func testAccEndpointConfigurationConfigAsyncKMSConfig(rName string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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

func testAccEndpointConfigurationConfigAsyncConfig(rName string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
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

  async_inference_config {
    output_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
    }
  }
}
`, rName)
}

func testAccEndpointConfigurationConfigAsyncNotifConfig(rName string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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

func testAccEndpointConfigurationConfigAsyncClientConfig(rName string) string {
	return testAccEndpointConfigurationConfig_Base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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
