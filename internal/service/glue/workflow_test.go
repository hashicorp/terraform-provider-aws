// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueWorkflow_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var workflow awstypes.Workflow

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "glue", fmt.Sprintf("workflow/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccGlueWorkflow_maxConcurrentRuns(t *testing.T) {
	ctx := acctest.Context(t)
	var workflow awstypes.Workflow

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_maxConcurrentRuns(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", "1"),
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
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", "2"),
				),
			},
			{
				Config: testAccWorkflowConfig_maxConcurrentRuns(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", "1"),
				),
			},
		},
	})
}

func TestAccGlueWorkflow_defaultRunProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var workflow awstypes.Workflow

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_defaultRunProperties(rName, "firstPropValue", "secondPropValue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "default_run_properties.%", "2"),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "First Description"),
				),
			},
			{
				Config: testAccWorkflowConfig_description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
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
				Config: testAccWorkflowConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWorkflowConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlueWorkflow_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var workflow awstypes.Workflow

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWorkflow(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					acctest.CheckSDKResourceDisappears(ctx, t, tfglue.ResourceWorkflow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckWorkflow(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)

	_, err := conn.ListWorkflows(ctx, &glue.ListWorkflowsInput{})

	// Some endpoints that do not support Glue Workflows return InternalFailure
	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckWorkflowExists(ctx context.Context, t *testing.T, resourceName string, workflow *awstypes.Workflow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Workflow ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)

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

func testAccCheckWorkflowDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_workflow" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)

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
