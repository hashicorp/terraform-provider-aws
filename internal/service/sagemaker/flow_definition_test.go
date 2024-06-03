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

func testAccFlowDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var flowDefinition sagemaker.DescribeFlowDefinitionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_flow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowDefinitionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowDefinitionExists(ctx, resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, "flow_definition_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("flow-definition/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "human_loop_request_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "human_loop_activation_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.public_workforce_task_price.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "human_loop_config.0.human_task_ui_arn", "aws_sagemaker_human_task_ui.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.task_availability_lifetime_in_seconds", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.task_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.task_description", rName),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.task_title", rName),
					resource.TestCheckResourceAttrPair(resourceName, "human_loop_config.0.workteam_arn", "aws_sagemaker_workteam.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "output_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "output_config.0.s3_output_path"),
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

func testAccFlowDefinition_humanLoopConfig_publicWorkforce(t *testing.T) {
	ctx := acctest.Context(t)
	var flowDefinition sagemaker.DescribeFlowDefinitionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_flow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowDefinitionConfig_publicWorkforce(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowDefinitionExists(ctx, resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, "flow_definition_name", rName),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.public_workforce_task_price.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.public_workforce_task_price.0.amount_in_usd.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.public_workforce_task_price.0.amount_in_usd.0.cents", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "human_loop_config.0.public_workforce_task_price.0.amount_in_usd.0.tenth_fractions_of_a_cent", acctest.Ct2),
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

func testAccFlowDefinition_humanLoopRequestSource(t *testing.T) {
	ctx := acctest.Context(t)
	var flowDefinition sagemaker.DescribeFlowDefinitionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_flow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowDefinitionConfig_humanLoopRequestSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowDefinitionExists(ctx, resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, "flow_definition_name", rName),
					resource.TestCheckResourceAttr(resourceName, "human_loop_request_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "human_loop_request_source.0.aws_managed_human_loop_request_source", "AWS/Textract/AnalyzeDocument/Forms/V1"),
					resource.TestCheckResourceAttr(resourceName, "human_loop_activation_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "human_loop_activation_config.0.human_loop_activation_conditions_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "human_loop_activation_config.0.human_loop_activation_conditions_config.0.human_loop_activation_conditions"),
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

func testAccFlowDefinition_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var flowDefinition sagemaker.DescribeFlowDefinitionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_flow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowDefinitionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowDefinitionExists(ctx, resourceName, &flowDefinition),
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
				Config: testAccFlowDefinitionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowDefinitionExists(ctx, resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFlowDefinitionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowDefinitionExists(ctx, resourceName, &flowDefinition),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccFlowDefinition_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var flowDefinition sagemaker.DescribeFlowDefinitionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_flow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowDefinitionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowDefinitionExists(ctx, resourceName, &flowDefinition),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceFlowDefinition(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceFlowDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFlowDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_flow_definition" {
				continue
			}

			_, err := tfsagemaker.FindFlowDefinitionByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker Flow Definition %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFlowDefinitionExists(ctx context.Context, n string, flowDefinition *sagemaker.DescribeFlowDefinitionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker Flow Definition ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		output, err := tfsagemaker.FindFlowDefinitionByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*flowDefinition = *output

		return nil
	}
}

func testAccFlowDefinitionBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_human_task_ui" "test" {
  human_task_ui_name = %[1]q

  ui_template {
    content = file("test-fixtures/sagemaker-human-task-ui-tmpl.html")
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccFlowDefinitionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFlowDefinitionBaseConfig(rName),
		testAccWorkteamConfig_cognito(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_flow_definition" "test" {
  flow_definition_name = %[1]q
  role_arn             = aws_iam_role.test.arn

  human_loop_config {
    human_task_ui_arn                     = aws_sagemaker_human_task_ui.test.arn
    task_availability_lifetime_in_seconds = 1
    task_count                            = 1
    task_description                      = %[1]q
    task_title                            = %[1]q
    workteam_arn                          = aws_sagemaker_workteam.test.arn
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }
}
`, rName))
}

func testAccFlowDefinitionConfig_publicWorkforce(rName string) string {
	return acctest.ConfigCompose(testAccFlowDefinitionBaseConfig(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_sagemaker_flow_definition" "test" {
  flow_definition_name = %[1]q
  role_arn             = aws_iam_role.test.arn

  human_loop_config {
    human_task_ui_arn                     = aws_sagemaker_human_task_ui.test.arn
    task_availability_lifetime_in_seconds = 1
    task_count                            = 1
    task_description                      = %[1]q
    task_title                            = %[1]q
    workteam_arn                          = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:394669845002:workteam/public-crowd/default"

    public_workforce_task_price {
      amount_in_usd {
        cents                     = 1
        tenth_fractions_of_a_cent = 2
      }
    }
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }
}
`, rName))
}

func testAccFlowDefinitionConfig_humanLoopRequestSource(rName string) string {
	return acctest.ConfigCompose(testAccFlowDefinitionBaseConfig(rName),
		testAccWorkteamConfig_cognito(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_flow_definition" "test" {
  flow_definition_name = %[1]q
  role_arn             = aws_iam_role.test.arn

  human_loop_config {
    human_task_ui_arn                     = aws_sagemaker_human_task_ui.test.arn
    task_availability_lifetime_in_seconds = 1
    task_count                            = 1
    task_description                      = %[1]q
    task_title                            = %[1]q
    workteam_arn                          = aws_sagemaker_workteam.test.arn
  }

  human_loop_request_source {
    aws_managed_human_loop_request_source = "AWS/Textract/AnalyzeDocument/Forms/V1"
  }

  human_loop_activation_config {
    human_loop_activation_conditions_config {
      human_loop_activation_conditions = <<EOF
        {
			"Conditions": [
			  {
				"ConditionType": "Sampling",
				"ConditionParameters": {
				  "RandomSamplingPercentage": 5
				}
			  }
			]
		}
        EOF
    }
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }
}
`, rName))
}

func testAccFlowDefinitionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccFlowDefinitionBaseConfig(rName),
		testAccWorkteamConfig_cognito(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_flow_definition" "test" {
  flow_definition_name = %[1]q
  role_arn             = aws_iam_role.test.arn

  human_loop_config {
    human_task_ui_arn                     = aws_sagemaker_human_task_ui.test.arn
    task_availability_lifetime_in_seconds = 1
    task_count                            = 1
    task_description                      = %[1]q
    task_title                            = %[1]q
    workteam_arn                          = aws_sagemaker_workteam.test.arn
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccFlowDefinitionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccFlowDefinitionBaseConfig(rName),
		testAccWorkteamConfig_cognito(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_flow_definition" "test" {
  flow_definition_name = %[1]q
  role_arn             = aws_iam_role.test.arn

  human_loop_config {
    human_task_ui_arn                     = aws_sagemaker_human_task_ui.test.arn
    task_availability_lifetime_in_seconds = 1
    task_count                            = 1
    task_description                      = %[1]q
    task_title                            = %[1]q
    workteam_arn                          = aws_sagemaker_workteam.test.arn
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
