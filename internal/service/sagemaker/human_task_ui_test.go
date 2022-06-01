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

func TestAccSageMakerHumanTaskUI_basic(t *testing.T) {
	var humanTaskUi sagemaker.DescribeHumanTaskUiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_human_task_ui.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHumanTaskUIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHumanTaskUIConfig_cognitoBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHumanTaskUIExists(resourceName, &humanTaskUi),
					resource.TestCheckResourceAttr(resourceName, "human_task_ui_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("human-task-ui/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "ui_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var humanTaskUi sagemaker.DescribeHumanTaskUiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_human_task_ui.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHumanTaskUIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHumanTaskUIConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHumanTaskUIExists(resourceName, &humanTaskUi),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ui_template.0.content", "ui_template.0.url"},
			},
			{
				Config: testAccHumanTaskUIConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHumanTaskUIExists(resourceName, &humanTaskUi),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccHumanTaskUIConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHumanTaskUIExists(resourceName, &humanTaskUi),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerHumanTaskUI_disappears(t *testing.T) {
	var humanTaskUi sagemaker.DescribeHumanTaskUiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_human_task_ui.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHumanTaskUIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHumanTaskUIConfig_cognitoBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHumanTaskUIExists(resourceName, &humanTaskUi),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceHumanTaskUI(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckHumanTaskUIDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_human_task_ui" {
			continue
		}

		_, err := tfsagemaker.FindHumanTaskUIByName(conn, rs.Primary.ID)

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

func testAccCheckHumanTaskUIExists(n string, humanTaskUi *sagemaker.DescribeHumanTaskUiOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker HumanTaskUi ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

		output, err := tfsagemaker.FindHumanTaskUIByName(conn, rs.Primary.ID)

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
