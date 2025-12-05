// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Env var gate to permit running this in environments with a valid target ARN
const envGlueInboundIntegrationTargetARN = "GLUE_INBOUND_INTEGRATION_TARGET_ARN"

func TestAccGlueInboundIntegration_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	if os.Getenv(envGlueInboundIntegrationTargetARN) == "" {
		t.Skipf("skipping, set %s to a valid SageMaker Lakehouse target ARN", envGlueInboundIntegrationTargetARN)
	}

	ctx := acctest.Context(t)
	var v awstypes.Integration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_inbound_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInboundIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlueInboundIntegrationConfig_basic(rName, os.Getenv(envGlueInboundIntegrationTargetARN)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInboundIntegrationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "integration_name", rName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccInboundIntegrationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccGlueInboundIntegration_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	if os.Getenv(envGlueInboundIntegrationTargetARN) == "" {
		t.Skipf("skipping, set %s to a valid SageMaker Lakehouse target ARN", envGlueInboundIntegrationTargetARN)
	}

	ctx := acctest.Context(t)
	var v awstypes.Integration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_inbound_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInboundIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlueInboundIntegrationConfig_basic(rName, os.Getenv(envGlueInboundIntegrationTargetARN)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInboundIntegrationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfglue.ResourceInboundIntegration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInboundIntegrationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_inbound_integration" {
				continue
			}

			_, err := tfglue.FindInboundIntegrationByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Glue Inbound Integration still exists: %s", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckInboundIntegrationExists(ctx context.Context, n string, v *awstypes.Integration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		out, err := tfglue.FindInboundIntegrationByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}
		*v = *out
		return nil
	}
}

func testAccInboundIntegrationImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}
		return rs.Primary.Attributes[names.AttrARN], nil
	}
}

func testAccGlueInboundIntegrationConfig_basic(rName, targetARN string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_dynamodb_table" "test" {
  name           = %q
  hash_key       = "pk"
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "pk"
    type = "S"
  }

  point_in_time_recovery { enabled = true }
}

resource "aws_glue_inbound_integration" "test" {
  integration_name = %q
  source_arn       = aws_dynamodb_table.test.arn
  target_arn       = %q
}
`, rName, rName, targetARN)
}
