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

func TestAccSageMakerHumanTaskUI_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var humanTaskUi sagemaker.DescribeHumanTaskUiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_human_task_ui.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHumanTaskUIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHumanTaskUIConfig_cognitoBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHumanTaskUIExists(ctx, resourceName, &humanTaskUi),
					resource.TestCheckResourceAttr(resourceName, "human_task_ui_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("human-task-ui/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "ui_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ui_template.0.content", "ui_template.0.url"},
			},
		},
	})
}

func TestAccSageMakerHumanTaskUI_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var humanTaskUi sagemaker.DescribeHumanTaskUiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_human_task_ui.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHumanTaskUIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHumanTaskUIConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHumanTaskUIExists(ctx, resourceName, &humanTaskUi),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ui_template.0.content", "ui_template.0.url"},
			},
			{
				Config: testAccHumanTaskUIConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHumanTaskUIExists(ctx, resourceName, &humanTaskUi),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccHumanTaskUIConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHumanTaskUIExists(ctx, resourceName, &humanTaskUi),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerHumanTaskUI_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var humanTaskUi sagemaker.DescribeHumanTaskUiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_human_task_ui.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHumanTaskUIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHumanTaskUIConfig_cognitoBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHumanTaskUIExists(ctx, resourceName, &humanTaskUi),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceHumanTaskUI(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckHumanTaskUIDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_human_task_ui" {
				continue
			}

			_, err := tfsagemaker.FindHumanTaskUIByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker HumanTaskUi %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckHumanTaskUIExists(ctx context.Context, n string, humanTaskUi *sagemaker.DescribeHumanTaskUiOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker HumanTaskUi ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		output, err := tfsagemaker.FindHumanTaskUIByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*humanTaskUi = *output

		return nil
	}
}

func testAccHumanTaskUIConfig_cognitoBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_human_task_ui" "test" {
  human_task_ui_name = %[1]q

  ui_template {
    content = file("test-fixtures/sagemaker-human-task-ui-tmpl.html")
  }
}
`, rName)
}

func testAccHumanTaskUIConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_human_task_ui" "test" {
  human_task_ui_name = %[1]q

  ui_template {
    content = file("test-fixtures/sagemaker-human-task-ui-tmpl.html")
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccHumanTaskUIConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_human_task_ui" "test" {
  human_task_ui_name = %[1]q

  ui_template {
    content = file("test-fixtures/sagemaker-human-task-ui-tmpl.html")
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
