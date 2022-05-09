package transfer_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccTransferWorkflow_basic(t *testing.T) {
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`workflow/.+`)),
					resource.TestCheckResourceAttr(resourceName, "steps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.type", "DELETE"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.0.source_file_location", "${original.file}"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.#", "0"),
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

func TestAccTransferWorkflow_onExecution(t *testing.T) {
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowOnExecConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`workflow/.+`)),
					resource.TestCheckResourceAttr(resourceName, "steps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.type", "DELETE"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.0.source_file_location", "${original.file}"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.type", "DELETE"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.delete_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.delete_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.delete_step_details.0.source_file_location", "${original.file}"),
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

func TestAccTransferWorkflow_description(t *testing.T) {
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowDescConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
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

func TestAccTransferWorkflow_tags(t *testing.T) {
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &conf),
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
				Config: testAccWorkflowConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkflowConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccTransferWorkflow_disappears(t *testing.T) {
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tftransfer.ResourceWorkflow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckWorkflowExists(n string, v *transfer.DescribedWorkflow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Workflow ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn

		output, err := tftransfer.FindWorkflowByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWorkflowDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_workflow" {
			continue
		}

		_, err := tftransfer.FindWorkflowByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Transfer Workflow %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccWorkflowBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_workflow" "test" {
  steps {
    delete_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}
`, rName)
}

func testAccWorkflowDescConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_workflow" "test" {
  description = %[1]q

  steps {
    delete_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}
`, rName)
}

func testAccWorkflowOnExecConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_workflow" "test" {
  steps {
    delete_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }

  on_exception_steps {
    delete_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}
`, rName)
}

func testAccWorkflowConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_transfer_workflow" "test" {
  steps {
    delete_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccWorkflowConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_transfer_workflow" "test" {
  steps {
    delete_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
