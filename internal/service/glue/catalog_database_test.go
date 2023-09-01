// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGlueCatalogDatabase_full(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, glue.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("database/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "location_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_database.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
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
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, glue.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_permission(rName, "ALTER"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
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
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, glue.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_target(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.catalog_id", "aws_glue_catalog_database.test2", "catalog_id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.database_name", "aws_glue_catalog_database.test2", "name"),
					resource.TestCheckResourceAttr(resourceName, "target_database.0.region", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccCatalogDatabaseConfig_targetLocation(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.catalog_id", "aws_glue_catalog_database.test2", "catalog_id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.database_name", "aws_glue_catalog_database.test2", "name"),
					resource.TestCheckResourceAttr(resourceName, "target_database.0.region", ""),
				),
			},
		},
	})
}

func TestAccGlueCatalogDatabase_targetDatabaseWithRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var providers []*schema.Provider
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, glue.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_targetWithRegion(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.catalog_id", "aws_glue_catalog_database.test2", "catalog_id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.database_name", "aws_glue_catalog_database.test2", "name"),
					resource.TestCheckResourceAttr(resourceName, "target_database.0.region", acctest.AlternateRegion()),
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

func TestAccGlueCatalogDatabase_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, glue.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_tags1(rName, "key1", "value1"),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccCatalogDatabaseConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config:  testAccCatalogDatabaseConfig_tags1(rName, "key2", "value2"),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGlueCatalogDatabase_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, glue.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceCatalogDatabase(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDatabaseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_catalog_database" {
				continue
			}

			catalogId, dbName, err := tfglue.ReadCatalogID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfglue.FindDatabaseByName(ctx, conn, catalogId, dbName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Glue Catalog Database %s still exists", rs.Primary.ID)
		}

		return nil
	}
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

func testAccCatalogDatabaseConfig_target(rName string) string {
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

func testAccCatalogDatabaseConfig_targetLocation(rName string) string {
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

func testAccCatalogDatabaseConfig_targetWithRegion(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q

  target_database {
    catalog_id    = aws_glue_catalog_database.test2.catalog_id
    database_name = aws_glue_catalog_database.test2.name
    region        = %[2]q
  }
}

resource "aws_glue_catalog_database" "test2" {
  provider = "awsalternate"

  name = "%[1]s-2"
}
`, rName, acctest.AlternateRegion()))
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

func testAccCatalogDatabaseConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCatalogDatabaseConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCheckCatalogDatabaseExists(ctx context.Context, name string) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn(ctx)
		_, err = tfglue.FindDatabaseByName(ctx, conn, catalogId, dbName)

		return err
	}
}
