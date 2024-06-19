package glue_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueCatalogTableOptimizer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var catalogTableOptimizer glue.GetTableOptimizerOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table_optimizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCatalogTableOptimizer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, resourceName, &catalogTableOptimizer),
					resource.TestCheckResourceAttr(resourceName, "catalog_id", "123456789012"),
					resource.TestCheckResourceAttr(resourceName, "database_name", "test_database"),
					resource.TestCheckResourceAttr(resourceName, "table_name", "test_table"),
					resource.TestCheckResourceAttr(resourceName, "type", "compaction"),
					resource.TestCheckResourceAttr(resourceName, "configuration.role_arn", "arn:aws:iam::123456789012:role/example-role"),
					resource.TestCheckResourceAttr(resourceName, "configuration.enabled", "true"),
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

func TestAccGlueCatalogTableOptimizer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var catalogTableOptimizer glue.GetTableOptimizerOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_table_optimizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCatalogTableOptimizer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceCatalogTableOptimizer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckCatalogTableOptimizer(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn(ctx)

	_, err := conn.BatchGetTableOptimizerWithContext(ctx, &glue.BatchGetTableOptimizerInput{
		Entries: []*glue.BatchGetTableOptimizerEntry{
			{
				CatalogId:    aws.String("dummy-catalog-id"),
				DatabaseName: aws.String("dummy-database-name"),
				TableName:    aws.String("dummy-table-name"),
				Type:         aws.String("dummy-type"),
			},
		},
	})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckCatalogTableOptimizerExists(ctx context.Context, resourceName string, catalogTableOptimizer *glue.GetTableOptimizerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Catalog Table Optimizer ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn(ctx)
		idParts := strings.Split(rs.Primary.ID, ":")
		if len(idParts) != 4 {
			return fmt.Errorf("unexpected format of ID (%q), expected catalog_id:database_name:table_name:type", rs.Primary.ID)
		}
		catalogID, databaseName, tableName, optimizerType := idParts[0], idParts[1], idParts[2], idParts[3]

		resp, err := conn.GetTableOptimizerWithContext(ctx, &glue.GetTableOptimizerInput{
			CatalogId:    aws.String(catalogID),
			DatabaseName: aws.String(databaseName),
			TableName:    aws.String(tableName),
			Type:         aws.String(optimizerType),
		})

		if err != nil {
			return fmt.Errorf("error getting Glue Catalog Table Optimizer (%s): %w", rs.Primary.ID, err)
		}

		if resp == nil || resp.TableOptimizer == nil {
			return fmt.Errorf("Glue Catalog Table Optimizer (%s) not found", rs.Primary.ID)
		}

		*catalogTableOptimizer = *resp

		return nil
	}
}

func testAccCheckCatalogTableOptimizerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_catalog_table_optimizer" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn(ctx)
			idParts := strings.Split(rs.Primary.ID, ":")
			if len(idParts) != 4 {
				return fmt.Errorf("unexpected format of ID (%q), expected catalog_id:database_name:table_name:type", rs.Primary.ID)
			}
			catalogID, databaseName, tableName, optimizerType := idParts[0], idParts[1], idParts[2], idParts[3]

			_, err := conn.GetTableOptimizerWithContext(ctx, &glue.GetTableOptimizerInput{
				CatalogId:    aws.String(catalogID),
				DatabaseName: aws.String(databaseName),
				TableName:    aws.String(tableName),
				Type:         aws.String(optimizerType),
			})
			if err != nil {
				if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
					return nil
				}
				return fmt.Errorf("error getting Glue Catalog Table Optimizer (%s): %w", rs.Primary.ID, err)
			}

			return fmt.Errorf("Glue Catalog Table Optimizer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCatalogTableOptimizerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_table_optimizer" "test" {
  catalog_id     = "123456789012"
  database_name  = "test_database"
  table_name     = "test_table"
  configuration  = {
    role_arn = "arn:aws:iam::123456789012:role/example-role"
    enabled  = true
  }
  type = "compaction"
}
`, rName)
}
