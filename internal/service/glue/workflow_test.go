// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueWorkflow_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var workflow awstypes.Workflow

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("workflow/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccGlueWorkflow_maxConcurrentRuns(t *testing.T) {
	ctx := acctest.Context(t)
	var workflow awstypes.Workflow

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_maxConcurrentRuns(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkflowConfig_maxConcurrentRuns(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", acctest.Ct2),
				),
			},
			{
				Config: testAccWorkflowConfig_maxConcurrentRuns(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccGlueWorkflow_defaultRunProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var workflow awstypes.Workflow

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_defaultRunProperties(rName, "firstPropValue", "secondPropValue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "default_run_properties.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_run_properties.--run-prop1", "firstPropValue"),
					resource.TestCheckResourceAttr(resourceName, "default_run_properties.--run-prop2", "secondPropValue"),
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

func TestAccGlueWorkflow_description(t *testing.T) {
	ctx := acctest.Context(t)
	var workflow awstypes.Workflow

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "First Description"),
				),
			},
			{
				Config: testAccWorkflowConfig_description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Second Description"),
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

func TestAccGlueWorkflow_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var workflow awstypes.Workflow
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
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
				Config: testAccWorkflowConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWorkflowConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlueWorkflow_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var workflow awstypes.Workflow

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, resourceName, &workflow),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceWorkflow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckWorkflow(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

	_, err := conn.ListWorkflows(ctx, &glue.ListWorkflowsInput{})

	// Some endpoints that do not support Glue Workflows return InternalFailure
	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckWorkflowExists(ctx context.Context, resourceName string, workflow *awstypes.Workflow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Workflow ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		output, err := conn.GetWorkflow(ctx, &glue.GetWorkflowInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output.Workflow == nil {
			return fmt.Errorf("Glue Workflow (%s) not found", rs.Primary.ID)
		}

		if aws.ToString(output.Workflow.Name) == rs.Primary.ID {
			*workflow = *output.Workflow
			return nil
		}

		return fmt.Errorf("Glue Workflow (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckWorkflowDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_workflow" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

			output, err := conn.GetWorkflow(ctx, &glue.GetWorkflowInput{
				Name: aws.String(rs.Primary.ID),
			})

			if err != nil {
				if errs.IsA[*awstypes.EntityNotFoundException](err) {
					return nil
				}
			}

			workflow := output.Workflow
			if workflow != nil && aws.ToString(workflow.Name) == rs.Primary.ID {
				return fmt.Errorf("Glue Workflow %s still exists", rs.Primary.ID)
			}

			return err
		}

		return nil
	}
}

func testAccWorkflowConfig_defaultRunProperties(rName, firstPropValue, secondPropValue string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = "%s"

  default_run_properties = {
    "--run-prop1" = "%s"
    "--run-prop2" = "%s"
  }
}
`, rName, firstPropValue, secondPropValue)
}

func testAccWorkflowConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  description = "%s"
  name        = "%s"
}
`, description, rName)
}

func testAccWorkflowConfig_required(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = "%s"
}
`, rName)
}

func testAccWorkflowConfig_maxConcurrentRuns(rName string, runs int) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name                = %[1]q
  max_concurrent_runs = %[2]d
}
`, rName, runs)
}

func testAccWorkflowConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccWorkflowConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
