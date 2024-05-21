// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerStudioLifecycleConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeStudioLifecycleConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_studio_lifecycle_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStudioLifecycleDestroyConfig(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStudioLifecycleConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioLifecycleExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "studio_lifecycle_config_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("studio-lifecycle-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "studio_lifecycle_config_app_type", "JupyterServer"),
					resource.TestCheckResourceAttrSet(resourceName, "studio_lifecycle_config_content"),
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

func TestAccSageMakerStudioLifecycleConfig_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeStudioLifecycleConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_studio_lifecycle_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStudioLifecycleDestroyConfig(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStudioLifecycleConfigConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioLifecycleExistsConfig(ctx, resourceName, &config),
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
				Config: testAccStudioLifecycleConfigConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioLifecycleExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStudioLifecycleConfigConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioLifecycleExistsConfig(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerStudioLifecycleConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var config sagemaker.DescribeStudioLifecycleConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_studio_lifecycle_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStudioLifecycleDestroyConfig(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStudioLifecycleConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioLifecycleExistsConfig(ctx, resourceName, &config),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceStudioLifecycleConfig(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceStudioLifecycleConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStudioLifecycleDestroyConfig(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_studio_lifecycle_config" {
				continue
			}

			_, err := tfsagemaker.FindStudioLifecycleConfigByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker Studio Lifecycle Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStudioLifecycleExistsConfig(ctx context.Context, n string, config *sagemaker.DescribeStudioLifecycleConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker Studio Lifecycle Config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		output, err := tfsagemaker.FindStudioLifecycleConfigByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*config = *output

		return nil
	}
}

func testAccStudioLifecycleConfigConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_studio_lifecycle_config" "test" {
  studio_lifecycle_config_name     = %[1]q
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")
}
`, rName)
}

func testAccStudioLifecycleConfigConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_studio_lifecycle_config" "test" {
  studio_lifecycle_config_name     = %[1]q
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccStudioLifecycleConfigConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_studio_lifecycle_config" "test" {
  studio_lifecycle_config_name     = %[1]q
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
