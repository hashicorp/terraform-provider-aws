// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package athena_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	glue "github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAthenaExecuteQueryAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	tableName := "test_table"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccExecuteQueryConfig_basic(rName, tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, t, tableName, rName),
				),
			},
		},
	})
}

func TestAccAthenaExecuteQueryAction_withInvalidQuery(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccExecuteQueryConfig_withInvalidQuery(rName),
				ExpectError: regexache.MustCompile("Query execution failed"),
			},
		},
	})
}

func testAccCheckTableExists(ctx context.Context, t *testing.T, tableName string, dbName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)
		input := &glue.GetTableInput{
			Name:         &tableName,
			DatabaseName: &dbName,
		}

		_, err := conn.GetTable(ctx, input)

		if err != nil {
			return fmt.Errorf("Error while getting %s table: %w", tableName, err)
		}

		return nil
	}
}

func testAccExecuteQueryConfig_basic(rName string, tableName string) string {
	return acctest.ConfigCompose(
		testAccExecuteActionConfig_base(rName),
		fmt.Sprintf(`
action "aws_athena_execute_query" "test" {
  config {
    query_string = <<EOT
CREATE EXTERNAL TABLE %[1]s (
    test INT
    )
LOCATION 's3://${aws_s3_bucket.test.id}/%[1]s/'
EOT
    workgroup = "primary"
    query_execution_context {
	  database = aws_glue_catalog_database.test.name
	}
	result_configuration {
  	  output_location = "s3://${aws_s3_bucket.test.id}/query-results/"
    }
  }
}
`, tableName))
}

func testAccExecuteQueryConfig_withInvalidQuery(rName string) string {
	return acctest.ConfigCompose(
		testAccExecuteActionConfig_base(rName),
		`
action "aws_athena_execute_query" "test" {
  config {
    query_string = "SELECT COUNT(*) FROM non_existent_table"
    workgroup    = "primary"
    query_execution_context {
	  database = aws_glue_catalog_database.test.name
	}
	result_configuration {
  	  output_location = "s3://${aws_s3_bucket.test.id}/query-results/"
    }
  }
}
`)
}

func testAccExecuteActionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  force_destroy = true
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_athena_execute_query.test]
    }
  }
}
`, rName)
}
