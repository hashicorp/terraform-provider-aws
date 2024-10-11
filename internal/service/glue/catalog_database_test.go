// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueCatalogDatabase_full(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("database/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "location_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_database.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "A test catalog from terraform"),
					resource.TestCheckResourceAttr(resourceName, "location_uri", "my-location"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "parameters.param2", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "parameters.param3", "50"),
				),
			},
			{
				Config: testAccCatalogDatabaseConfig_full(rName, "An updated test catalog from terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "An updated test catalog from terraform"),
					resource.TestCheckResourceAttr(resourceName, "location_uri", "my-location"),
					resource.TestCheckResourceAttr(resourceName, "parameters.param1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "parameters.param2", acctest.CtTrue),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_permission(rName, "ALTER"),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.0.permissions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "create_table_default_permission.0.permissions.*", "ALTER"),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.0.principal.#", acctest.Ct1),
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
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.0.permissions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "create_table_default_permission.0.permissions.*", "SELECT"),
					resource.TestCheckResourceAttr(resourceName, "create_table_default_permission.0.principal.#", acctest.Ct1),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_target(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_database.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.catalog_id", "aws_glue_catalog_database.test2", names.AttrCatalogID),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.database_name", "aws_glue_catalog_database.test2", names.AttrName),
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
					resource.TestCheckResourceAttr(resourceName, "target_database.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.catalog_id", "aws_glue_catalog_database.test2", names.AttrCatalogID),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.database_name", "aws_glue_catalog_database.test2", names.AttrName),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_targetWithRegion(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_database.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.catalog_id", "aws_glue_catalog_database.test2", names.AttrCatalogID),
					resource.TestCheckResourceAttrPair(resourceName, "target_database.0.database_name", "aws_glue_catalog_database.test2", names.AttrName),
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

func TestAccGlueCatalogDatabase_federatedDatabase(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_catalog_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_federatedDatabase(rName),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("database/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "federated_database.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "federated_database.0.connection_name", "aws:redshift"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "federated_database.0.identifier", "redshift", regexache.MustCompile(`datashare:+.`)),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccCatalogDatabaseConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccCatalogDatabaseConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config:  testAccCatalogDatabaseConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCatalogDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

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

func testAccCatalogDatabaseConfig_federatedDatabase(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
  db_name        = "test"
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftdata_statement" "test_create" {
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
  database       = aws_redshiftserverless_namespace.test.db_name
  sql            = "CREATE DATASHARE tfacctest;"
}
`, rName),
		// Split this resource into a string literal so the terraform `format` function
		// interpolates properly
		`
resource "aws_redshiftdata_statement" "test_grant_usage" {
  depends_on     = [aws_redshiftdata_statement.test_create]
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
  database       = aws_redshiftserverless_namespace.test.db_name
  sql            = format("GRANT USAGE ON DATASHARE tfacctest TO ACCOUNT '%s' VIA DATA CATALOG;", data.aws_caller_identity.current.account_id)
}

locals {
  # Data share ARN is not returned from the GRANT USAGE statement, so must be
  # composed manually.
  # Ref: https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonredshift.html#amazonredshift-resources-for-iam-policies
  data_share_arn = format("arn:%s:redshift:%s:%s:datashare:%s/%s",
    data.aws_partition.current.id,
    data.aws_region.current.name,
    data.aws_caller_identity.current.account_id,
    aws_redshiftserverless_namespace.test.namespace_id,
    "tfacctest",
  )
}

resource "aws_redshift_data_share_authorization" "test" {
  depends_on = [aws_redshiftdata_statement.test_grant_usage]

  data_share_arn      = local.data_share_arn
  consumer_identifier = format("DataCatalog/%s", data.aws_caller_identity.current.account_id)
}

resource "aws_redshift_data_share_consumer_association" "test" {
  depends_on = [aws_redshift_data_share_authorization.test]

  data_share_arn = local.data_share_arn
  consumer_arn = format("arn:%s:glue:%s:%s:catalog",
    data.aws_partition.current.id,
    data.aws_region.current.name,
    data.aws_caller_identity.current.account_id,
  )
}

resource "aws_lakeformation_resource" "test" {
  depends_on = [aws_redshift_data_share_consumer_association.test]

  arn                     = local.data_share_arn
  use_service_linked_role = false
}
`,
		fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  depends_on = [aws_lakeformation_resource.test]

  name = %[1]q
  federated_database {
    connection_name = "aws:redshift"
    identifier      = local.data_share_arn
  }
}
`, rName))
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)
		_, err = tfglue.FindDatabaseByName(ctx, conn, catalogId, dbName)

		return err
	}
}
