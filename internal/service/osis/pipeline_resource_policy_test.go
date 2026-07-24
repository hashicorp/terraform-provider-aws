// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfosis "github.com/hashicorp/terraform-provider-aws/internal/service/osis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchIngestionPipelineResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	pipelineName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, acctest.RandString(t, 10))
	resourceName := "aws_osis_pipeline_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchIngestionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineResourcePolicyConfig_basic(pipelineName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineResourcePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrResourceARN),
			},
		},
	})
}

func TestAccOpenSearchIngestionPipelineResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, acctest.RandString(t, 10))
	resourceName := "aws_osis_pipeline_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchIngestionEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchIngestionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineResourcePolicyExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfosis.ResourcePipelineResourcePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccOpenSearchIngestionPipelineResourcePolicy_disappears_pipeline(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, acctest.RandString(t, 10))
	resourceName := "aws_osis_pipeline_resource_policy.test"
	pipelineResourceName := "aws_osis_pipeline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchIngestionEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchIngestionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineResourcePolicyExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfosis.ResourcePipeline, pipelineResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOpenSearchIngestionPipelineResourcePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, acctest.RandString(t, 10))
	resourceName := "aws_osis_pipeline_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchIngestionEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchIngestionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineResourcePolicyExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccPipelineResourcePolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineResourcePolicyExists(ctx, t, resourceName),
					acctest.CheckResourceAttrJMES(resourceName, names.AttrPolicy, "Statement[0].Sid", "AllowIngestUpdated"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckPipelineResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchIngestionClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_osis_pipeline_resource_policy" {
				continue
			}

			resourceArn := rs.Primary.Attributes[names.AttrResourceARN]

			_, err := tfosis.FindPipelineResourcePolicyByResourceARN(ctx, conn, resourceArn)

			if err == nil {
				return fmt.Errorf("OpenSearch Ingestion Pipeline Resource Policy (%s) still exists", resourceArn)
			}

			if !retry.NotFound(err) {
				return err
			}
		}

		return nil
	}
}

func testAccCheckPipelineResourcePolicyExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		resourceArn := rs.Primary.Attributes[names.AttrResourceARN]
		if resourceArn == "" {
			return fmt.Errorf("No OpenSearch Ingestion Pipeline Resource Policy resource_arn is set")
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchIngestionClient(ctx)

		_, err := tfosis.FindPipelineResourcePolicyByResourceARN(ctx, conn, resourceArn)

		return err
	}
}

func testAccPipelineResourcePolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPipelineConfig_basic(rName), `
data "aws_caller_identity" "current" {}

resource "aws_osis_pipeline_resource_policy" "test" {
  resource_arn = aws_osis_pipeline.test.pipeline_arn

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "ingestPolicy",
  "Statement": [{
    "Sid": "AllowIngest",
    "Effect": "Allow",
    "Principal": {
      "AWS": "${data.aws_caller_identity.current.account_id}"
    },
    "Action": [
      "osis:CreatePipelineEndpoint"
    ],
    "Resource": "${aws_osis_pipeline.test.pipeline_arn}"
  }]
}
EOF
}
`)
}

func testAccPipelineResourcePolicyConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccPipelineConfig_basic(rName), `
data "aws_caller_identity" "current" {}

resource "aws_osis_pipeline_resource_policy" "test" {
  resource_arn = aws_osis_pipeline.test.pipeline_arn

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowIngestUpdated",
    "Effect": "Allow",
    "Principal": {
      "AWS": "*"
    },
    "Action": [
      "osis:CreatePipelineEndpoint"
    ],
    "Resource": "${aws_osis_pipeline.test.pipeline_arn}"
  }]
}
EOF
}
`)
}
