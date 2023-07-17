// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccTransferWorkflow_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`workflow/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.custom_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.decrypt_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.0.source_file_location", "${original.file}"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.tag_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.type", "DELETE"),
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

func TestAccTransferWorkflow_onExceptionSteps(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_onExceptionSteps(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`workflow/.+`)),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.copy_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.custom_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.decrypt_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.type", "DELETE"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.delete_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.delete_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.delete_step_details.0.source_file_location", "${original.file}"),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.0.tag_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.custom_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.decrypt_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.0.source_file_location", "${original.file}"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.tag_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.type", "DELETE"),
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
	ctx := acctest.Context(t)
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_description(rName, "testing"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "testing"),
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
	ctx := acctest.Context(t)
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &conf),
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
				Config: testAccWorkflowConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkflowConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccTransferWorkflow_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceWorkflow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTransferWorkflow_allSteps(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedWorkflow
	resourceName := "aws_transfer_workflow.test"
	rName := sdkacctest.RandString(25)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_allSteps(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`workflow/.+`)),
					resource.TestCheckResourceAttr(resourceName, "on_exception_steps.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.0.destination_file_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.0.destination_file_location.0.efs_file_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.0.destination_file_location.0.s3_file_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.0.destination_file_location.0.s3_file_location.0.bucket", "testing"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.0.destination_file_location.0.s3_file_location.0.key", "k1"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.0.overwrite_existing", "TRUE"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.copy_step_details.0.source_file_location", "${original.file}"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.custom_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.decrypt_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.delete_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.tag_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.0.type", "COPY"),
					resource.TestCheckResourceAttr(resourceName, "steps.1.copy_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.1.custom_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.1.custom_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "steps.1.custom_step_details.0.source_file_location", "${original.file}"),
					resource.TestCheckResourceAttrPair(resourceName, "steps.1.custom_step_details.0.target", "aws_lambda_function.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "steps.1.custom_step_details.0.timeout_seconds", "1001"),
					resource.TestCheckResourceAttr(resourceName, "steps.1.decrypt_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.1.delete_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.1.tag_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.1.type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.copy_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.custom_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.decrypt_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.decrypt_step_details.0.destination_file_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.decrypt_step_details.0.destination_file_location.0.efs_file_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "steps.2.decrypt_step_details.0.destination_file_location.0.efs_file_location.0.file_system_id", "aws_efs_file_system.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.decrypt_step_details.0.destination_file_location.0.efs_file_location.0.path", "/test"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.decrypt_step_details.0.destination_file_location.0.s3_file_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.decrypt_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "steps.2.decrypt_step_details.0.overwrite_existing", "FALSE"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.decrypt_step_details.0.source_file_location", "${original.file}"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.decrypt_step_details.0.type", "PGP"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.delete_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.tag_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.2.type", "DECRYPT"),
					resource.TestCheckResourceAttr(resourceName, "steps.3.copy_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.3.custom_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.3.decrypt_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.3.delete_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.3.delete_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "steps.3.delete_step_details.0.source_file_location", "${original.file}"),
					resource.TestCheckResourceAttr(resourceName, "steps.3.tag_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.3.type", "DELETE"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.copy_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.custom_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.decrypt_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.delete_step_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.tag_step_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.tag_step_details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "steps.4.tag_step_details.0.source_file_location", "${original.file}"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.tag_step_details.0.tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.tag_step_details.0.tags.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.tag_step_details.0.tags.0.value", "value1"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.tag_step_details.0.tags.1.key", "key2"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.tag_step_details.0.tags.1.value", "value2"),
					resource.TestCheckResourceAttr(resourceName, "steps.4.type", "TAG"),
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

func testAccCheckWorkflowExists(ctx context.Context, n string, v *transfer.DescribedWorkflow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Workflow ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		output, err := tftransfer.FindWorkflowByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWorkflowDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_workflow" {
				continue
			}

			_, err := tftransfer.FindWorkflowByID(ctx, conn, rs.Primary.ID)

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
}

func testAccWorkflowConfig_basic(rName string) string {
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

func testAccWorkflowConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_transfer_workflow" "test" {
  description = %[2]q

  steps {
    delete_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}
`, rName, description)
}

func testAccWorkflowConfig_onExceptionSteps(rName string) string {
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

func testAccWorkflowConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

func testAccWorkflowConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccWorkflowConfig_allSteps(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs14.x"
}

resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_transfer_workflow" "test" {
  steps {
    copy_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
      destination_file_location {
        s3_file_location {
          bucket = "testing"
          key    = "k1"
        }
      }
      overwrite_existing = "TRUE"
    }
    type = "COPY"
  }

  steps {
    custom_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
      target               = aws_lambda_function.test.arn
      timeout_seconds      = 1001
    }
    type = "CUSTOM"
  }

  steps {
    decrypt_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
      destination_file_location {
        efs_file_location {
          file_system_id = aws_efs_file_system.test.id
          path           = "/test"
        }
      }
      type = "PGP"
    }
    type = "DECRYPT"
  }

  steps {
    delete_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }

  steps {
    tag_step_details {
      name                 = %[1]q
      source_file_location = "$${original.file}"

      tags {
        key   = "key1"
        value = "value1"
      }

      tags {
        key   = "key2"
        value = "value2"
      }
    }
    type = "TAG"
  }
}
`, rName))
}
