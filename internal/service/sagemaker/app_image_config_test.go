package sagemaker_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
)

func TestAccSageMakerAppImageConfig_basic(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppImageDestroyConfig,
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("app-image-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "0"),
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

func TestAccSageMakerAppImageConfig_KernelGatewayImage_kernelSpecs(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppImageDestroyConfig,
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageKernelGatewayImageKernalSpecs1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppImageKernelGatewayImageKernalSpecs2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.display_name", rName),
				),
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_KernelGatewayImage_fileSystem(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppImageDestroyConfig,
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageKernelGatewayImageFileSystem1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_gid", "100"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_uid", "1000"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.mount_path", "/home/sagemaker-user"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppImageKernelGatewayImageFileSystem2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_gid", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_uid", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.mount_path", "/test"),
				),
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_tags(t *testing.T) {
	var app sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(resourceName, &app),
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
				Config: testAccAppImageTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppImageTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_disappears(t *testing.T) {
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppImageDestroyConfig,
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(resourceName, &config),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceAppImageConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppImageDestroyConfig(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_app_image_config" {
			continue
		}

		config, err := tfsagemaker.FindAppImageConfigByName(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading SageMaker App Image Config (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(config.AppImageConfigName) == rs.Primary.ID {
			return fmt.Errorf("SageMaker App Image Config %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAppImageExistsConfig(n string, config *sagemaker.DescribeAppImageConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker App Image Config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		resp, err := tfsagemaker.FindAppImageConfigByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*config = *resp

		return nil
	}
}

func testAccAppImageBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q
}
`, rName)
}

func testAccAppImageKernelGatewayImageKernalSpecs1Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }
  }
}
`, rName)
}

func testAccAppImageKernelGatewayImageKernalSpecs2Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name         = "%[1]s-2"
      display_name = %[1]q
    }
  }
}
`, rName)
}

func testAccAppImageKernelGatewayImageFileSystem1Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }

    file_system_config {}
  }
}
`, rName)
}

func testAccAppImageKernelGatewayImageFileSystem2Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  kernel_gateway_image_config {
    kernel_spec {
      name = %[1]q
    }

    file_system_config {
      default_gid = 0
      default_uid = 0
      mount_path  = "/test"
    }
  }
}
`, rName)
}

func testAccAppImageTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppImageTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
