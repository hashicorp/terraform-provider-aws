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
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerNotebookInstanceLifecycleConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var lifecycleConfig sagemaker.DescribeNotebookInstanceLifecycleConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance_lifecycle_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceLifecycleConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceLifecycleConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceLifecycleConfigurationExists(ctx, t, resourceName, &lifecycleConfig),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckNoResourceAttr(resourceName, "on_create"),
					resource.TestCheckNoResourceAttr(resourceName, "on_start"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("notebook-instance-lifecycle-config/%s", rName)),
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

func TestAccSageMakerNotebookInstanceLifecycleConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var lifecycleConfig sagemaker.DescribeNotebookInstanceLifecycleConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance_lifecycle_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceLifecycleConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceLifecycleConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceLifecycleConfigurationExists(ctx, t, resourceName, &lifecycleConfig),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccNotebookInstanceLifecycleConfigurationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceLifecycleConfigurationExists(ctx, t, resourceName, &lifecycleConfig),

					resource.TestCheckResourceAttr(resourceName, "on_create", inttypes.Base64EncodeOnce([]byte("echo bla"))),
					resource.TestCheckResourceAttr(resourceName, "on_start", inttypes.Base64EncodeOnce([]byte("echo blub"))),
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

func TestAccSageMakerNotebookInstanceLifecycleConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var lifecycleConfig sagemaker.DescribeNotebookInstanceLifecycleConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance_lifecycle_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceLifecycleConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceLifecycleConfigurationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceLifecycleConfigurationExists(ctx, t, resourceName, &lifecycleConfig),
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
				Config: testAccNotebookInstanceLifecycleConfigurationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceLifecycleConfigurationExists(ctx, t, resourceName, &lifecycleConfig),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccNotebookInstanceLifecycleConfigurationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceLifecycleConfigurationExists(ctx, t, resourceName, &lifecycleConfig),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerNotebookInstanceLifecycleConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var lifecycleConfig sagemaker.DescribeNotebookInstanceLifecycleConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_notebook_instance_lifecycle_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotebookInstanceLifecycleConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNotebookInstanceLifecycleConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotebookInstanceLifecycleConfigurationExists(ctx, t, resourceName, &lifecycleConfig),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsagemaker.ResourceNotebookInstanceLifeCycleConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNotebookInstanceLifecycleConfigurationExists(ctx context.Context, t *testing.T, resourceName string, lifecycleConfig *sagemaker.DescribeNotebookInstanceLifecycleConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		output, err := tfsagemaker.FindNotebookInstanceLifecycleConfigByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*lifecycleConfig = *output

		return nil
	}
}

func testAccCheckNotebookInstanceLifecycleConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_notebook_instance_lifecycle_configuration" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

			_, err := tfsagemaker.FindNotebookInstanceLifecycleConfigByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker AI Notebook Instance Lifecycle Configuration %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccNotebookInstanceLifecycleConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "test" {
  name = %q
}
`, rName)
}

func testAccNotebookInstanceLifecycleConfigurationConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "test" {
  name      = %q
  on_create = base64encode("echo bla")
  on_start  = base64encode("echo blub")
}
`, rName)
}

func testAccNotebookInstanceLifecycleConfigurationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccNotebookInstanceLifecycleConfigurationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
