// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerAppImageConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("app-image-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_gateway_image_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccSageMakerAppImageConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &config),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsagemaker.ResourceAppImageConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_KernelGatewayImage_kernelSpecs(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_kernelGatewayKernalSpecs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &config),
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
				Config: testAccAppImageConfigConfig_kernelGatewayKernalSpecs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &config),
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
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_kernelGatewayFileSystem1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &config),
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
				Config: testAccAppImageConfigConfig_kernelGatewayFileSystem2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &config),
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

func TestAccSageMakerAppImageConfig_CodeEditor(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_codeEditor(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_arguments.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_arguments.0", rName),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_entrypoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_entrypoint.0", rName),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_environment_variables.%", "1"),
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
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_arguments.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_arguments.0", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_entrypoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_entrypoint.0", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "code_editor_app_image_config.0.container_config.0.container_environment_variables.%", "1"),
				),
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_JupyterLab(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_jupyterLab(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_arguments.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_arguments.0", rName),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_entrypoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_entrypoint.0", rName),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_environment_variables.%", "1"),
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
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "app_image_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_arguments.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_arguments.0", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_entrypoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_entrypoint.0", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "jupyter_lab_image_config.0.container_config.0.container_environment_variables.%", "1"),
				),
			},
		},
	})
}

func TestAccSageMakerAppImageConfig_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var app sagemaker.DescribeAppImageConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_app_image_config.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppImageConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppImageConfigConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAppImageConfigConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppImageConfigExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckAppImageConfigDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_app_image_config" {
				continue
			}

			_, err := tfsagemaker.FindAppImageConfigByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker AI App Image Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppImageConfigExists(ctx context.Context, t *testing.T, n string, v *sagemaker.DescribeAppImageConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		output, err := tfsagemaker.FindAppImageConfigByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAppImageConfigConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = %[1]q

  code_editor_app_image_config {}
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

  code_editor_app_image_config {}

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

  code_editor_app_image_config {}

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
