// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/osis"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfosis "github.com/hashicorp/terraform-provider-aws/internal/service/osis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOSISPipelineResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	pipelineName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, sdkacctest.RandString(10))
	resourceName := "aws_osis_pipeline_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchIngestionEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineResourcePolicyConfig_basic(pipelineName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineResourcePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "resource_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
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

func TestAccOSISPipelineResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceArn := "aws_osis_pipeline_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchIngestionEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchIngestionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineResourcePolicyExists(ctx, resourceArn),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfosis.ResourcePipelineResourcePolicy, resourceArn),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOSISPipelineResourcePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_osis_pipeline_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchIngestionEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchIngestionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineResourcePolicyExists(ctx, resourceName),
				),
			},
			{
				Config: testAccPipelineResourcePolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineResourcePolicyExists(ctx, resourceName),
				),
			},
		},
	})
}

func testAccCheckPipelineResourcePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchIngestionClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_osis_pipeline_resource_policy" {
				continue
			}

			output, err := conn.GetResourcePolicy(ctx, &osis.GetResourcePolicyInput{
				ResourceArn: &rs.Primary.ID,
			})
			if *output.Policy != "{}" {
				return fmt.Errorf("OpenSearch Ingestion Pipeline Resource Policy (%s) still exists", rs.Primary.ID)
			}

			if err != nil {
				return fmt.Errorf("error getting OpenSearch Ingestion pipeline Resource Policy (%s) to verify destroy: %w", rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccCheckPipelineResourcePolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchIngestionClient(ctx)

		_, err := conn.GetResourcePolicy(ctx, &osis.GetResourcePolicyInput{
			ResourceArn: &rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("Error reading OpenSearch Ingestion Pipeline Resource Policy (%s): %w", rs.Primary.ID, err)
		}

		return nil
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
	return acctest.ConfigCompose(fmt.Sprintf(testAccPipelineConfig_basic(rName), `

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
      "AWS": "${data.aws_caller_identity.target.account_id}"
    },
    "Action": [
      "osis:CreatePipelineEndpoint"
    ],
    "Resource": "${aws_osis_pipeline.test.pipeline_arn}"
  }]
}
EOF
}
`))
}
