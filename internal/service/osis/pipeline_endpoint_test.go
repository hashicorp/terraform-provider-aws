// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/osis/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfosis "github.com/hashicorp/terraform-provider-aws/internal/service/osis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchIngestionPipelineEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pipelineEndpoint awstypes.PipelineEndpoint
	rName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, acctest.RandString(t, 10))
	resourceName := "aws_osis_pipeline_endpoint.test"
	pipelineResourceName := "aws_osis_pipeline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchIngestionEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchIngestionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineEndpointExists(ctx, t, resourceName, &pipelineEndpoint),
					resource.TestCheckResourceAttrPair(resourceName, "pipeline_arn", pipelineResourceName, "pipeline_arn"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.0.security_group_ids.#", "1"),
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

func TestAccOpenSearchIngestionPipelineEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var pipelineEndpoint awstypes.PipelineEndpoint
	rName := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, acctest.RandString(t, 10))
	resourceName := "aws_osis_pipeline_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchIngestionEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchIngestionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineEndpointExists(ctx, t, resourceName, &pipelineEndpoint),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfosis.ResourcePipelineEndpoint, resourceName),
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

func testAccCheckPipelineEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchIngestionClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_osis_pipeline_endpoint" {
				continue
			}

			_, err := tfosis.FindPipelineEndpointByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Ingestion Pipeline Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPipelineEndpointExists(ctx context.Context, t *testing.T, n string, v *awstypes.PipelineEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No OpenSearch Ingestion Pipeline Endpoint ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchIngestionClient(ctx)

		output, err := tfosis.FindPipelineEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPipelineEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPipelineEndpointConfig_pipeline(rName), fmt.Sprintf(`
resource "aws_osis_pipeline_endpoint" "test" {
  pipeline_arn = aws_osis_pipeline.test.pipeline_arn

  vpc_options {
    subnet_ids         = [aws_subnet.test.id]
    security_group_ids = [aws_security_group.test.id]
  }
}

resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  name   = "%[1]s-endpoint"
  vpc_id = aws_vpc.test.id
}
`, rName))
}

func testAccPipelineEndpointConfig_pipeline(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "osis-pipelines.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_vpc" "pipeline" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "pipeline" {
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.pipeline.id
}

resource "aws_security_group" "pipeline" {
  name   = %[1]q
  vpc_id = aws_vpc.pipeline.id
}

resource "aws_osis_pipeline" "test" {
  pipeline_name               = %[1]q
  pipeline_configuration_body = <<-EOT
            version: "2"
            test-pipeline:
              source:
                http:
                  path: "/test"
              sink:
                - s3:
                    aws:
                      sts_role_arn: "${aws_iam_role.test.arn}"
                      region: "${data.aws_region.current.region}"
                    bucket: "test"
                    threshold:
                      event_collect_timeout: "60s"
                    codec:
                      ndjson:
        EOT
  max_units                   = 1
  min_units                   = 1

  vpc_options {
    security_group_ids      = [aws_security_group.pipeline.id]
    subnet_ids              = [aws_subnet.pipeline.id]
    vpc_endpoint_management = "SERVICE"
  }
}
`, rName)
}
