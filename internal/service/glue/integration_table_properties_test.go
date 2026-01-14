// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const envGlueITPTargetTableARN = "GLUE_ITP_TARGET_TABLE_ARN"

func TestAccGlueIntegrationTableProperties_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	if os.Getenv(envGlueITPTargetTableARN) == "" {
		t.Skipf("skipping, set %s to a Glue Catalog target table ARN", envGlueITPTargetTableARN)
	}

	ctx := acctest.Context(t)
	resourceName := "aws_glue_integration_table_properties.test"
	ttArn := os.Getenv(envGlueITPTargetTableARN)
	ttName := testAccGlueITPExtractTableName(ttArn)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationTablePropertiesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlueIntegrationTablePropertiesConfig_basic(ttArn, ttName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceARN, ttArn),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, ttName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccITPImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
			},
		},
	})
}

func testAccGlueITPExtractTableName(arn string) string {
	// ARN format: arn:aws:glue:region:account:table/dbname/tablename
	// Extract substring after last '/'
	for i := len(arn) - 1; i >= 0; i-- {
		if arn[i] == '/' {
			return arn[i+1:]
		}
	}
	return arn
}

func testAccCheckIntegrationTablePropertiesDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_integration_table_properties" {
				continue
			}
			_, err := tfglue.FindIntegrationTableProperties(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN], rs.Primary.Attributes[names.AttrTableName])
			if err == nil {
				return fmt.Errorf("Glue Integration Table Properties still exists: %s,%s", rs.Primary.Attributes[names.AttrResourceARN], rs.Primary.Attributes[names.AttrTableName])
			}
		}
		return nil
	}
}

func testAccITPImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}
		return fmt.Sprintf("%s,%s", rs.Primary.Attributes[names.AttrResourceARN], rs.Primary.Attributes[names.AttrTableName]), nil
	}
}

func testAccGlueIntegrationTablePropertiesConfig_basic(targetTableARN, tableName string) string {
	return fmt.Sprintf(`
resource "aws_glue_integration_table_properties" "test" {
  resource_arn = %[1]q
  table_name   = %[2]q

  target_table_config {
    target_table_name = %[2]q
    unnest_spec       = "TOPLEVEL"

    partition_spec {
      field_name      = "event_time"
      function_spec   = "hour"
      conversion_spec = "epoch_milli"
    }
  }
}
`, targetTableARN, tableName)
}
