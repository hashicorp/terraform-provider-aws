// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueFederatedCatalogDataSource_basic(t *testing.T) {
	t.Skip("Skipping basic test - test-identifier is not a registered data source")
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_glue_federated_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFederatedCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFederatedCatalogDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, "Test federated catalog"),
					resource.TestCheckResourceAttr(dataSourceName, "federated_catalog.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "federated_catalog.0.identifier", "test-identifier"),
					resource.TestCheckResourceAttr(dataSourceName, "federated_catalog.0.connection_name", "test-connection"),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "glue", regexache.MustCompile(`catalog/.+$`)),
				),
			},
		},
	})
}

func TestAccGlueFederatedCatalogDataSource_s3Tables(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_glue_federated_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFederatedCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFederatedCatalogDataSourceConfig_s3Tables(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, "s3tablescatalog"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, "Test S3 Tables federated catalog"),
					resource.TestCheckResourceAttr(dataSourceName, "federated_catalog.#", "1"),
					resource.TestCheckResourceAttrSet(dataSourceName, "federated_catalog.0.identifier"),
					resource.TestCheckResourceAttr(dataSourceName, "federated_catalog.0.connection_name", "aws:s3tables"),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "glue", regexache.MustCompile(`catalog/.+$`)),
				),
			},
		},
	})
}

func testAccFederatedCatalogDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_federated_catalog" "test" {
  name        = %[1]q
  description = "Test federated catalog"

  federated_catalog {
    identifier      = "test-identifier"
    connection_name = "test-connection"
  }
}

data "aws_glue_federated_catalog" "test" {
  name = aws_glue_federated_catalog.test.name
}
`, rName)
}

func testAccFederatedCatalogDataSourceConfig_s3Tables(rName string) string {
	return acctest.ConfigCompose(
		testAccFederatedCatalogConfig_s3TablesBase(rName), `
resource "aws_glue_federated_catalog" "test" {
  name        = "s3tablescatalog"
  description = "Test S3 Tables federated catalog"

  federated_catalog {
    identifier      = "arn:${data.aws_partition.current.partition}:s3tables:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:bucket/*"
    connection_name = "aws:s3tables"
  }

  depends_on = [aws_lakeformation_resource.test]
}

data "aws_glue_federated_catalog" "test" {
  name = aws_glue_federated_catalog.test.name
}
`,
	)
}
