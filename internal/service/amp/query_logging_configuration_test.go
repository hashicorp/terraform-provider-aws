// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/amp/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAMPQueryLoggingConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.QueryLoggingConfigurationMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_query_logging_configuration.test"
	workspaceResourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueryLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.cloudwatch_logs.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "destination.0.cloudwatch_logs.0.log_group_arn"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.filters.0.qsp_threshold", "500"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "workspace_id"),
				ImportStateVerifyIdentifierAttribute: "workspace_id",
			},
		},
	})
}

func TestAccAMPQueryLoggingConfiguration_withFilters(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.QueryLoggingConfigurationMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_query_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLoggingConfigurationConfig_withFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.filters.0.qsp_threshold", "1000"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "workspace_id"),
				ImportStateVerifyIdentifierAttribute: "workspace_id",
			},
		},
	})
}

func TestAccAMPQueryLoggingConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.QueryLoggingConfigurationMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_query_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryLoggingConfigurationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfamp.ResourceQueryLoggingConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQueryLoggingConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_prometheus_query_logging_configuration" {
				continue
			}

			_, err := tfamp.FindQueryLoggingConfigurationByID(ctx, conn, rs.Primary.Attributes["workspace_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Prometheus Query Logging Configuration %s still exists", rs.Primary.Attributes["workspace_id"])
		}

		return nil
	}
}

func testAccCheckQueryLoggingConfigurationExists(ctx context.Context, n string, v *types.QueryLoggingConfigurationMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPClient(ctx)

		output, err := tfamp.FindQueryLoggingConfigurationByID(ctx, conn, rs.Primary.Attributes["workspace_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccQueryLoggingConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias = %[1]q
}

resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/prometheus/query-logs/%[1]s"
}

resource "aws_prometheus_query_logging_configuration" "test" {
  workspace_id = aws_prometheus_workspace.test.id

  destination {
    cloudwatch_logs {
      log_group_arn = "${aws_cloudwatch_log_group.test.arn}:*"
    }

    filters {
      qsp_threshold = 500
    }
  }
}
`, rName)
}

func testAccQueryLoggingConfigurationConfig_withFilters(rName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias = %[1]q
}

resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/prometheus/query-logs/%[1]s-custom"
}

resource "aws_prometheus_query_logging_configuration" "test" {
  workspace_id = aws_prometheus_workspace.test.id

  destination {
    cloudwatch_logs {
      log_group_arn = "${aws_cloudwatch_log_group.test.arn}:*"
    }

    filters {
      qsp_threshold = 1000
    }
  }
}
`, rName)
}
