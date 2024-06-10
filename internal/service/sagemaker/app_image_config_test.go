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

func TestAccSageMakerAppImageConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageDestroyConfig(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("app-image-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageDestroyConfig(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_kernelGatewayKernalSpecs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppImageConfigConfig_kernelGatewayKernalSpecs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.0.display_name", rName),
				),
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_KernelGatewayImage_fileSystem(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageDestroyConfig(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_kernelGatewayFileSystem1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", acctest.Ct1),
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
				Config: testAccAppImageConfigConfig_kernelGatewayFileSystem2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.kernel_spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_gid", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.default_uid", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.0.file_system_config.0.mount_path", "/test"),
				),
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_CodeEditor(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageDestroyConfig(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_codeEditor(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_arguments.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_arguments.0", rName),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_entrypoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_entrypoint.0", rName),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_environment_variables.%", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppImageConfigConfig_codeEditor(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_arguments.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_arguments.0", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_entrypoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_entrypoint.0", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_environment_variables.%", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_JupyterLab(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageDestroyConfig(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_jupyterLab(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_arguments.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_arguments.0", rName),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_entrypoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_entrypoint.0", rName),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_environment_variables.%", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppImageConfigConfig_jupyterLab(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_arguments.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_arguments.0", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_entrypoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_entrypoint.0", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_environment_variables.%", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var app sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppImageConfigConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAppImageConfigConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageDestroyConfig(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageExistsConfig(ctx, resourceName, &config),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceAppImageConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppImageDestroyConfig(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_app_image_config" {
				continue
			}

			config, err := tfsagemaker.FindAppImageConfigByName(ctx, conn, rs.Primary.ID)

			if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker App Image Config (%s): %w", rs.Primary.ID, err)
			}

			if aws.StringValue(config.AppImageConfigName) == rs.Primary.ID {
				return fmt.Errorf("SageMaker App Image Config %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckAppImageExistsConfig(ctx context.Context, n string, config *sagemaker.DescribeAppImageConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker App Image Config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)
		resp, err := tfsagemaker.FindAppImageConfigByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*config = *resp

		return nil
	}
}

func testAccAppImageConfigConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q
}
`, rName)
}

func testAccAppImageConfigConfig_kernelGatewayKernalSpecs1(rName string) string {
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

func testAccAppImageConfigConfig_kernelGatewayKernalSpecs2(rName string) string {
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

func testAccAppImageConfigConfig_kernelGatewayFileSystem1(rName string) string {
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

func testAccAppImageConfigConfig_kernelGatewayFileSystem2(rName string) string {
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

func testAccAppImageConfigConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppImageConfigConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccAppImageConfigConfig_jupyterLab(rName, arg string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  jupyter_lab_image_config {
    container_config {
      container_arguments  = ["%[2]s"]
      container_entrypoint = ["%[2]s"]
      container_environment_variables = {
        %[2]q = %[2]q
      }
    }
  }
}
`, rName, arg)
}

func testAccAppImageConfigConfig_codeEditor(rName, arg string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  code_editor_app_image_config {
    container_config {
      container_arguments  = ["%[2]s"]
      container_entrypoint = ["%[2]s"]
      container_environment_variables = {
        %[2]q = %[2]q
      }
    }
  }
}
`, rName, arg)
}
