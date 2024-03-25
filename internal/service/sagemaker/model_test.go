// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerModel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_container.0.image", "data.aws_sagemaker_prebuilt_ecr_image.test", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.mode", "SingleModel"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.environment.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", "arn"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("model/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enable_network_isolation", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "inference_execution_config.#", "0"),
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

func TestAccSageMakerModel_inferenceExecution(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_inferenceExecution(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "inference_execution_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_execution_config.0.mode", "Serial"),
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

func TestAccSageMakerModel_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccModelConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccModelConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func TestAccSageMakerModel_primaryContainerModelDataURL(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerDataURL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "primary_container.0.model_data_url"),
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

func TestAccSageMakerModel_primaryContainerHostname(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerHostname(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.container_hostname", "test"),
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

func TestAccSageMakerModel_primaryContainerImage(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerImage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image_config.0.repository_access_mode", "Platform"),
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

func TestAccSageMakerModel_primaryContainerEnvironment(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerEnvironment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.environment.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.environment.test", "bar"),
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

func TestAccSageMakerModel_primaryContainerModeSingle(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerModeSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.mode", "SingleModel"),
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

func TestAccSageMakerModel_primaryContainerModelPackageName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerPackageName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "primary_container.0.model_package_name"),
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

func TestAccSageMakerModel_primaryContainerModelDataSource(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerUncompressedModel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.model_data_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.model_data_source.0.s3_data_source.0.s3_data_type", "S3Prefix"),
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

func TestAccSageMakerModel_containers(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_containers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "container.0.image", "data.aws_sagemaker_prebuilt_ecr_image.test", "registry_path"),
					resource.TestCheckResourceAttrPair(resourceName, "container.1.image", "data.aws_sagemaker_prebuilt_ecr_image.test", "registry_path"),
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

func TestAccSageMakerModel_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_vpcBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "2"),
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

func TestAccSageMakerModel_primaryContainerPrivateDockerRegistry(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerPrivateDockerRegistry(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image_config.0.repository_access_mode", "Vpc"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image_config.0.repository_auth_config.0.repository_credentials_provider_arn", "arn:aws:lambda:us-east-2:123456789012:function:my-function:1"), //lintignore:AWSAT003,AWSAT005
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

func TestAccSageMakerModel_networkIsolation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_networkIsolation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_network_isolation", "true"),
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

func TestAccSageMakerModel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceModel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckModelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_model" {
				continue
			}

			resp, err := conn.ListModelsWithContext(ctx, &sagemaker.ListModelsInput{
				NameContains: aws.String(rs.Primary.ID),
			})
			if err == nil {
				if len(resp.Models) > 0 {
					return fmt.Errorf("SageMaker models still exist")
				}

				return nil
			}

			if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
				return err
			}
		}

		return nil
	}
}

func testAccCheckModelExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker model ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)
		DescribeModelOpts := &sagemaker.DescribeModelInput{
			ModelName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribeModelWithContext(ctx, DescribeModelOpts)

		return err
	}
}

func testAccModel_primaryContainerModelS3DataSourceAcceptEula(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerModelS3DataSourceAcceptEula(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.model_data_source.0.s3_data_source.0.model_access_config.0.accept_eula", "true"),
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

func testAccModel_primaryContainerInferenceSpecificationName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerInferenceSpecificationName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.inference_specification_name", "test"),
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

func testAccModel_primaryContainerMultiModelConfigModelCacheSetting(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_primaryContainerMultiModelConfigModelCacheSetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.multi_model_config.0.model_cache_setting", "Disabled"),
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

func testAccModel_containersMultiModelConfigModelCacheSetting(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelConfig_containersMultiModelConfigModelCacheSetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "container.0.multi_model_config.0.model_cache_setting", "Disabled"),
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

func testAccModelConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "kmeans"
}
`, rName)
}

func testAccModelConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}
`, rName))
}

func testAccModelConfig_inferenceExecution(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  inference_execution_config {
    mode = "Serial"
  }

  container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}
`, rName))
}

func testAccModelConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccModelConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccModelConfig_primaryContainerDataURL(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image          = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    model_data_url = "https://s3.amazonaws.com/${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  }
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  description = "Allow SageMaker to create model"
  policy      = data.aws_iam_policy_document.policy.json
}

data "aws_iam_policy_document" "policy" {
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
    ]

    resources = [
      "*",
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:GetObject",
    ]

    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "model.tar.gz"
  content = "some-data"
}
`, rName))
}

// lintignore:AWSAT003,AWSAT005
func testAccModelConfig_primaryContainerPackageName(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
data "aws_region" "current" {}

locals {
  region_account_map = {
    us-east-1      = "865070037744"
    us-east-2      = "057799348421"
    us-west-1      = "382657785993"
    us-west-2      = "594846645681"
    ca-central-1   = "470592106596"
    eu-central-1   = "446921602837"
    eu-west-1      = "985815980388"
    eu-west-2      = "856760150666"
    eu-west-3      = "843114510376"
    eu-north-1     = "136758871317"
    ap-southeast-1 = "192199979996"
    ap-southeast-2 = "666831318237"
    ap-northeast-2 = "745090734665"
    ap-northeast-1 = "977537786026"
    ap-south-1     = "077584701553"
    sa-east-1      = "270155090741"
  }

  account = local.region_account_map[data.aws_region.current.name]

  model_package_name = format(
    "arn:aws:sagemaker:%%s:%%s:model-package/hf-textgeneration-gpt2-cpu-b73b575105d336b680d151277ebe4ee0",
    data.aws_region.current.name,
    local.account
  )
}

resource "aws_sagemaker_model" "test" {
  name                     = %[1]q
  enable_network_isolation = true
  execution_role_arn       = aws_iam_role.test.arn

  primary_container {
    model_package_name = local.model_package_name
  }
}
`, rName))
}

func testAccModelConfig_primaryContainerHostname(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image              = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    container_hostname = "test"
  }
}
`, rName))
}

func testAccModelConfig_primaryContainerImage(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path

    image_config {
      repository_access_mode = "Platform"
    }
  }
}
`, rName))
}

func testAccModelConfig_primaryContainerEnvironment(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path

    environment = {
      test = "bar"
    }
  }
}
`, rName))
}

func testAccModelConfig_primaryContainerModeSingle(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    mode  = "SingleModel"
  }
}
`, rName))
}

func testAccModelConfig_primaryContainerUncompressedModel(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    model_data_source {
      s3_data_source {
        s3_data_type     = "S3Prefix"
        s3_uri           = "s3://${aws_s3_object.test.bucket}/model/"
        compression_type = "None"
      }
    }
  }
}


resource "aws_iam_policy" "test" {
  name        = %[1]q
  description = "Allow SageMaker to create model"
  policy      = data.aws_iam_policy_document.policy.json
}

data "aws_iam_policy_document" "policy" {
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
    ]

    resources = [
      "*",
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:GetObject",
      "s3:ListBucket",
    ]

    resources = [
      "${aws_s3_bucket.test.arn}",
      "${aws_s3_bucket.test.arn}/*",
    ]
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "model/inference.py"
  content = "some-data"
}
`, rName))
}

func testAccModelConfig_containers(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}
`, rName))
}

func testAccModelConfig_networkIsolation(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name                     = %[1]q
  execution_role_arn       = aws_iam_role.test.arn
  enable_network_isolation = true

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}
`, rName))
}

func testAccModelConfig_vpcBasic(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name                     = %[1]q
  execution_role_arn       = aws_iam_role.test.arn
  enable_network_isolation = true

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  vpc_config {
    subnets            = aws_subnet.test[*].id
    security_group_ids = aws_security_group.test[*].id
  }
}

resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

// lintignore:AWSAT003,AWSAT005
func testAccModelConfig_primaryContainerPrivateDockerRegistry(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name                     = %[1]q
  execution_role_arn       = aws_iam_role.test.arn
  enable_network_isolation = true

  primary_container {
    image = "registry.example.com/test-model"

    image_config {
      repository_access_mode = "Vpc"

      repository_auth_config {
        repository_credentials_provider_arn = "arn:aws:lambda:us-east-2:123456789012:function:my-function:1"
      }
    }
  }

  vpc_config {
    subnets            = aws_subnet.test[*].id
    security_group_ids = [aws_security_group.test.id]
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccModelConfig_primaryContainerModelS3DataSourceAcceptEula(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}
locals {
  region_account_map = {
    us-east-1      = "763104351884"
    us-east-2      = "763104351884"
    us-west-1      = "763104351884"
    us-west-2      = "763104351884"
    af-south-1     = "626614931356"
    ca-central-1   = "763104351884"
    eu-central-1   = "763104351884"
    eu-central-2   = "380420809688"
    eu-west-1      = "763104351884"
    eu-west-2      = "763104351884"
    eu-west-3      = "763104351884"
    eu-north-1     = "763104351884"
    eu-south-1     = "692866216735"
    eu-south-2     = "503227376785"
    il-central-1   = "780543022126"
    me-south-1     = "217643126080"
    me-central-1   = "914824155844"
    sa-east-1      = "763104351884"
    ap-southeast-1 = "763104351884"
    ap-southeast-2 = "763104351884"
    ap-southeast-3 = "907027046896"
    ap-southeast-4 = "457447274322"
    ap-northeast-1 = "763104351884"
    ap-northeast-2 = "763104351884"
    ap-northeast-3 = "364406365360"
    ap-south-1     = "763104351884"
    ap-south-2     = "772153158452"
    ap-east-1      = "871362719292"
    sa-east-1      = "763104351884"
    cn-north-1     = "727897471807"
    cn-northwest-1 = "727897471807"
  }
  account = local.region_account_map[data.aws_region.current.name]
  primary_container_image = format(
    "%%s.dkr.ecr.%%s.%%s/huggingface-pytorch-tgi-inference:2.0.1-tgi1.1.0-gpu-py39-cu118-ubuntu20.04",
    local.account,
    data.aws_region.current.name,
    data.aws_partition.current.dns_suffix
    
  primary_container_model_data_source_s3_uri = format(
    "s3://jumpstart-private-cache-prod-%%s/meta-textgeneration/meta-textgeneration-llama-2-13b-f/artifacts/inference-prepack/v1.0.0/",
    data.aws_region.current.name
  )
}

resource "aws_sagemaker_model" "test" {
  name                     = %[1]q
  enable_network_isolation = true
  execution_role_arn       = aws_iam_role.test.arn
  
  primary_container {
    image = local.primary_container_image
    mode = "SingleModel"
    environment = {
      ENDPOINT_SERVER_TIMEOUT = 3600
      HF_MODEL_ID = "/opt/ml/model"
      MAX_INPUT_LENGTH = 4095
      MAX_TOTAL_TOKENS = 4096
      MODEL_CACHE_ROOT = "/opt/ml/model"
      SAGEMAKER_ENV = 1
      SAGEMAKER_MODEL_SERVER_WORKERS = 1
      SAGEMAKER_PROGRAM = "inference.py"
      SM_NUM_GPUS = 4
  }

  model_data_source {
    s3_data_source {
      compression_type = "None"
      s3_data_type     = "S3Prefix"
      s3_uri           = local.primary_container_model_data_source_s3_uri
      model_access_config {
        accept_eula = true
      }
    }
  }
}
`, rName))
}

func testAccModelConfig_primaryContainerInferenceSpecificationName(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`

resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image                        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    inference_specification_name = "test"
  }
}
`, rName))
}

func testAccModelConfig_primaryContainerMultiModelConfigModelCacheSetting(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}
locals {
  region_account_map = {
    "us-east-1"      = "785573368785"
    "us-east-2"      = "007439368137"
    "us-west-1"      = "710691900526"
    "us-west-2"      = "301217895009"
    "eu-west-1"      = "802834080501"
    "eu-west-2"      = "205493899709"
    "eu-west-3"      = "254080097072"
    "eu-north-1"     = "601324751636"
    "eu-south-1"     = "966458181534"
    "eu-central-1"   = "746233611703"
    "ap-east-1"      = "110948597952"
    "ap-south-1"     = "763008648453"
    "ap-northeast-1" = "941853720454"
    "ap-northeast-2" = "151534178276"
    "ap-southeast-1" = "324986816169"
    "ap-southeast-2" = "355873309152"
    "cn-northwest-1" = "474822919863"
    "cn-north-1"     = "472730292857"
    "sa-east-1"      = "756306329178"
    "ca-central-1"   = "464438896020"
    "me-south-1"     = "836785723513"
    "af-south-1"     = "774647643957"
  }
  account = local.region_account_map[data.aws_region.current.name]
  primary_container_image = format(
    "%%s.dkr.ecr.%%s.%%s/sagemaker-tritonserver:22.07-py3",
    local.account,
    data.aws_region.current.name,
    data.aws_partition.current.dns_suffix
  )
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "resnet50-mme-gpu/model.tar.gz"
  content = "some-data"
}

resource "aws_sagemaker_model" "test" {
  depends_on = [aws_s3_object.test]

  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image          = local.primary_container_image
    mode           = "MultiModel"
    model_data_url = "s3://${aws_s3_bucket.test.id}/resnet50-mme-gpu/"
    multi_model_config {
      model_cache_setting = "Disabled"
    }
  }
}
`, rName))
}

func testAccModelConfig_containersMultiModelConfigModelCacheSetting(rName string) string {
	return acctest.ConfigCompose(testAccModelConfig_base(rName), fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}
locals {
  region_account_map = {
    "us-east-1"      = "785573368785"
    "us-east-2"      = "007439368137"
    "us-west-1"      = "710691900526"
    "us-west-2"      = "301217895009"
    "eu-west-1"      = "802834080501"
    "eu-west-2"      = "205493899709"
    "eu-west-3"      = "254080097072"
    "eu-north-1"     = "601324751636"
    "eu-south-1"     = "966458181534"
    "eu-central-1"   = "746233611703"
    "ap-east-1"      = "110948597952"
    "ap-south-1"     = "763008648453"
    "ap-northeast-1" = "941853720454"
    "ap-northeast-2" = "151534178276"
    "ap-southeast-1" = "324986816169"
    "ap-southeast-2" = "355873309152"
    "cn-northwest-1" = "474822919863"
    "cn-north-1"     = "472730292857"
    "sa-east-1"      = "756306329178"
    "ca-central-1"   = "464438896020"
    "me-south-1"     = "836785723513"
    "af-south-1"     = "774647643957"
  }
  account = local.region_account_map[data.aws_region.current.name]
  container_image = format(
    "%%s.dkr.ecr.%%s.%%s/sagemaker-tritonserver:22.07-py3",
    local.account,
    data.aws_region.current.name,
    data.aws_partition.current.dns_suffix
  )
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "resnet50-mme-gpu/model.tar.gz"
  content = "some-data"
}

resource "aws_sagemaker_model" "test" {
  depends_on = [aws_s3_object.test]

  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  container {
    image          = local.container_image
    mode           = "MultiModel"
    model_data_url = "s3://${aws_s3_bucket.test.id}/resnet50-mme-gpu/"
    multi_model_config {
      model_cache_setting = "Disabled"
    }
  }
}
`, rName))
}
