package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
)

func TestAccGlueCatalogDatabase_full(t *testing.T) {
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("database/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "location_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_database.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccCatalogDatabaseConfig_full(rName, "A test catalog from terraform"),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "A test catalog from terraform"),
					resource.TestCheckResourceAttr(resourceName, "location_uri", "my-location"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param2", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param3", "50"),
				),
			},
			{
				Config: testAccCatalogDatabaseConfig_full(rName, "An updated test catalog from terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "An updated test catalog from terraform"),
					resource.TestCheckResourceAttr(resourceName, "location_uri", "my-location"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param2", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param3", "50"),
				),
			},
		},
	})
}

func TestAccGlueCatalogDatabase_createTablePermission(t *testing.T) {
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_permission(rName, "ALTER"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.0.permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "create_table_default_permission.0.permissions.*", "ALTER"),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.0.principal.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.0.principal.0.data_lake_principal_identifier", "IAM_ALLOWED_PRINCIPALS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccCatalogDatabaseConfig_permission(rName, "SELECT"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.0.permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "create_table_default_permission.0.permissions.*", "SELECT"),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.0.principal.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.0.principal.0.data_lake_principal_identifier", "IAM_ALLOWED_PRINCIPALS"),
				),
			},
		},
	})
}

func TestAccGlueCatalogDatabase_targetDatabase(t *testing.T) {
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_targetDatabase(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.catalog_id", "aws_glue_catalog_database.test2", "catalog_id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.database_name", "aws_glue_catalog_database.test2", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccCatalogDatabaseConfig_targetDatabaseLocation(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.catalog_id", "aws_glue_catalog_database.test2", "catalog_id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.database_name", "aws_glue_catalog_database.test2", "name"),
				),
			},
		},
	})
}

func TestAccGlueCatalogDatabase_disappears(t *testing.T) {
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceCatalogDatabase(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDatabaseDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_catalog_database" {
			continue
		}

		catalogId, dbName, err := tfglue.ReadCatalogID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &glue.GetDatabaseInput{
			CatalogId: aws.String(catalogId),
			Name:      aws.String(dbName),
		}
		if _, err := conn.GetDatabase(input); err != nil {
			//Verify the error is what we want
			if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
				continue
			}

			return err
		}
		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccCatalogDatabaseConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}
`, rName)
}

func testAccCatalogDatabaseConfig_full(rName, desc string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name         = %[1]q
  description  = %[2]q
  location_uri = "my-location"

  parameters = {
    param1 = "value1"
    param2 = true
    param3 = 50
  }
}
`, rName, desc)
}

func testAccCatalogDatabaseConfig_targetDatabase(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q

  target_database {
    catalog_id    = aws_glue_catalog_database.test2.catalog_id
    database_name = aws_glue_catalog_database.test2.name
  }
}

resource "aws_glue_catalog_database" "test2" {
  name = "%[1]s-2"
}
`, rName)
}

func testAccCatalogDatabaseConfig_targetDatabaseLocation(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q

  target_database {
    catalog_id    = aws_glue_catalog_database.test2.catalog_id
    database_name = aws_glue_catalog_database.test2.name
  }
}

resource "aws_glue_catalog_database" "test2" {
  name         = "%[1]s-2"
  location_uri = "my-location"
}
`, rName)
}

func testAccCatalogDatabaseConfig_permission(rName, permission string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q

  create_table_default_permission {
    permissions = [%[2]q]

    principal {
      data_lake_principal_identifier = "IAM_ALLOWED_PRINCIPALS"
    }
  }
}
`, rName, permission)
}

func testAccCheckCatalogDatabaseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		catalogId, dbName, err := tfglue.ReadCatalogID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn
		out, err := conn.GetDatabase(&glue.GetDatabaseInput{
			CatalogId: aws.String(catalogId),
			Name:      aws.String(dbName),
		})

		if err != nil {
			return err
		}

		if out.Database == nil {
			return fmt.Errorf("No Glue Database Found")
		}

		if aws.StringValue(out.Database.Name) != dbName {
			return fmt.Errorf("Glue Database Mismatch - existing: %q, state: %q",
				aws.StringValue(out.Database.Name), dbName)
		}

		return nil
	}
}
