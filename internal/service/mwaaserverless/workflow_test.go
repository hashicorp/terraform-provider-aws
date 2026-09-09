// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mwaaserverless_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mwaaserverless"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmwaaserverless "github.com/hashicorp/terraform-provider-aws/internal/service/mwaaserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMWAAServerlessWorkflow_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var workflow mwaaserverless.GetWorkflowOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mwaaserverless_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MWAAServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Workflow/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "airflow-serverless", regexache.MustCompile("workflow/.+")),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "definition_s3_location.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "workflow_version"),
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

func TestAccMWAAServerlessWorkflow_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var workflow mwaaserverless.GetWorkflowOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mwaaserverless_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MWAAServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Workflow/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmwaaserverless.ResourceWorkflow, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccMWAAServerlessWorkflow_code(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var workflow mwaaserverless.GetWorkflowOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mwaaserverless_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MWAAServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_code(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "code.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "code.0.s3_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "code.0.s3_location.0.bucket", "aws_s3_bucket.test", names.AttrID),
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

func testAccCheckWorkflowDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MWAAServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mwaaserverless_workflow" {
				continue
			}

			_, err := tfmwaaserverless.FindWorkflowByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.MWAAServerless, create.ErrActionCheckingDestroyed, tfmwaaserverless.ResNameWorkflow, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckWorkflowExists(ctx context.Context, t *testing.T, name string, workflow *mwaaserverless.GetWorkflowOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MWAAServerless, create.ErrActionCheckingExistence, tfmwaaserverless.ResNameWorkflow, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MWAAServerless, create.ErrActionCheckingExistence, tfmwaaserverless.ResNameWorkflow, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).MWAAServerlessClient(ctx)

		output, err := tfmwaaserverless.FindWorkflowByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.MWAAServerless, create.ErrActionCheckingExistence, tfmwaaserverless.ResNameWorkflow, rs.Primary.ID, err)
		}

		*workflow = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).MWAAServerlessClient(ctx)

	input := &mwaaserverless.ListWorkflowsInput{}

	_, err := conn.ListWorkflows(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccWorkflowConfig_code(rName string) string {
	return fmt.Sprintf(`
resource "aws_mwaaserverless_workflow" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  definition_s3_location {
    bucket     = aws_s3_bucket.test.id
    object_key = aws_s3_object.definition.key
  }

  code {
    s3_location {
      bucket     = aws_s3_bucket.test.id
      object_key = aws_s3_object.code.key
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "definition" {
  bucket = aws_s3_bucket.test.id
  key    = "workflow.yaml"

  content = <<-YAML
    name: %[1]s
    schedule: "@daily"
    tasks:
      - id: start
        operator: airflow.operators.empty.EmptyOperator
  YAML
}

resource "aws_s3_object" "code" {
  bucket  = aws_s3_bucket.test.id
  key     = "code.txt"
  content = "code artifact placeholder"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "airflow-serverless.amazonaws.com"
      }
    }]
  })
}
`, rName)
}
