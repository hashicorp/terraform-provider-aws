// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSGlobalClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_global_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDatabaseName, resourceName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDeletionProtection, resourceName, names.AttrDeletionProtection),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngineVersion, resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIdentifier, resourceName, "global_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "members", resourceName, "global_cluster_members"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceID, resourceName, "global_cluster_resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStorageEncrypted, resourceName, names.AttrStorageEncrypted),
				),
			},
		},
	})
}

func testAccGlobalClusterDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_global_cluster" "test" {
  identifier = aws_rds_global_cluster.test.global_cluster_identifier
}

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "aurora-postgresql"
  database_name             = "example_db"
}
`, rName)
}
