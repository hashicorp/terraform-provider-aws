// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueIntegrationTableProperties_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var integrationTableProperties glue.GetIntegrationTablePropertiesOutput
	resourceName := "aws_glue_integration_table_properties.test"
	databaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	targetTableName := sdkacctest.RandString(10)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationTablePropertiesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationTablePropertiesConfig_basic(databaseName, tableName, targetTableName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntegrationTablePropertiesExists(ctx, t, resourceName, &integrationTableProperties),
					resource.TestCheckResourceAttr(resourceName, "table_name", tableName),
					resource.TestCheckResourceAttr(resourceName, "target_table_config.0.unnest_spec", "FULL"),
					resource.TestCheckResourceAttr(resourceName, "target_table_config.0.target_table_name", targetTableName),
					resource.TestCheckResourceAttr(resourceName, "target_table_config.0.partition_spec.0.field_name", "created_at"),
					resource.TestCheckResourceAttr(resourceName, "target_table_config.0.partition_spec.0.function_spec", "month"),
					resource.TestCheckResourceAttr(resourceName, "target_table_config.0.partition_spec.0.conversion_spec", "iso"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
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

func TestAccGlueIntegrationTableProperties_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var integrationTableProperties glue.GetIntegrationTablePropertiesOutput
	resourceName := "aws_glue_integration_table_properties.test"
	databaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	targetTableName := sdkacctest.RandString(10)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationTablePropertiesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationTablePropertiesConfig_basic(databaseName, tableName, targetTableName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntegrationTablePropertiesExists(ctx, t, resourceName, &integrationTableProperties),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfglue.ResourceIntegrationTableProperties, resourceName),
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

func testAccCheckIntegrationTablePropertiesDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_integration_table_properties" {
				continue
			}

			resourceArn, tableName, found := strings.Cut(rs.Primary.ID, intflex.ResourceIdSeparator)
			if !found {
				return create.Error(names.Glue, create.ErrActionCheckingDestroyed, tfglue.ResNameIntegrationTableProperties, rs.Primary.ID, errors.New("invalid ID format"))
			}

			_, err := tfglue.FindIntegrationTableProperties(ctx, conn, resourceArn, tableName)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Glue, create.ErrActionCheckingDestroyed, tfglue.ResNameIntegrationTableProperties, rs.Primary.ID, err)
			}

			return create.Error(names.Glue, create.ErrActionCheckingDestroyed, tfglue.ResNameIntegrationTableProperties, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckIntegrationTablePropertiesExists(ctx context.Context, t *testing.T, name string, integrationTableProperties *glue.GetIntegrationTablePropertiesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Glue, create.ErrActionCheckingExistence, tfglue.ResNameIntegrationTableProperties, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Glue, create.ErrActionCheckingExistence, tfglue.ResNameIntegrationTableProperties, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)

		resourceArn, tableName, found := strings.Cut(rs.Primary.ID, intflex.ResourceIdSeparator)
		if !found {
			return create.Error(names.Glue, create.ErrActionCheckingExistence, tfglue.ResNameIntegrationTableProperties, rs.Primary.ID, errors.New("invalid ID format"))
		}

		resp, err := tfglue.FindIntegrationTableProperties(ctx, conn, resourceArn, tableName)
		if err != nil {
			return create.Error(names.Glue, create.ErrActionCheckingExistence, tfglue.ResNameIntegrationTableProperties, rs.Primary.ID, err)
		}

		*integrationTableProperties = *resp

		return nil
	}
}

func testAccIntegrationTablePropertiesConfig_basic(databaseName, tableName, targetTableName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_integration_table_properties" "test" {
  resource_arn = aws_glue_catalog_database.test.arn
  table_name   = %[2]q

  target_table_config {
    unnest_spec       = "FULL"
    target_table_name = %[3]q

    partition_spec {
      field_name      = "created_at"
      function_spec   = "month"
      conversion_spec = "iso"
    }
  }
}
`, databaseName, tableName, targetTableName)
}
