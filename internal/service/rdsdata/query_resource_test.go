// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rdsdata_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSDataQueryResource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rdsdata_query.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSDataServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "records", "[{\"1\":1}]"),
					resource.TestCheckResourceAttr(resourceName, "number_of_records_updated", "0"),
				),
			},
		},
	})
}

func TestAccRDSDataQueryResource_withParameters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rdsdata_query.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSDataServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryResourceConfig_withParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "records"),
					resource.TestCheckResourceAttr(resourceName, "number_of_records_updated", "0"),
				),
			},
		},
	})
}

func testAccQueryResourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccQueryDataSourceConfig_base(rName), `
resource "aws_rdsdata_query" "test" {
  depends_on   = [aws_rds_cluster_instance.test]
  resource_arn = aws_rds_cluster.test.arn
  secret_arn   = aws_secretsmanager_secret_version.test.arn
  sql          = "SELECT 1"
}
`)
}

func testAccQueryResourceConfig_withParameters(rName string) string {
	return acctest.ConfigCompose(testAccQueryDataSourceConfig_base(rName), `
resource "aws_rdsdata_query" "test" {
  depends_on   = [aws_rds_cluster_instance.test]
  resource_arn = aws_rds_cluster.test.arn
  secret_arn   = aws_secretsmanager_secret_version.test.arn
  sql          = "SELECT * FROM information_schema.tables WHERE table_name = :table_name"
  database     = aws_rds_cluster.test.database_name

  parameters {
    name  = "table_name"
    value = "test_table"
  }
}
`)
}
