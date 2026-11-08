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

func TestAccRDSGlobalClustersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_global_clusters.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClustersDataSourceConfig_basic(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "global_cluster_identifiers.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "global_cluster_arns.#"),
				),
			},
		},
	})
}

func TestAccRDSGlobalClustersDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_global_clusters.test"
	resourceName := "aws_rds_global_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClustersDataSourceConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "global_cluster_identifiers.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_identifiers.0", resourceName, "global_cluster_identifier"),
				),
			},
		},
	})
}

func testAccGlobalClustersDataSourceConfig_basic(rName1, rName2 string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "aurora-postgresql"
}

resource "aws_rds_global_cluster" "test1" {
  global_cluster_identifier = %[1]q
  engine                    = data.aws_rds_engine_version.default.engine
  engine_version            = data.aws_rds_engine_version.default.version
}

resource "aws_rds_global_cluster" "test2" {
  global_cluster_identifier = %[2]q
  engine                    = data.aws_rds_engine_version.default.engine
  engine_version            = data.aws_rds_engine_version.default.version
}

data "aws_rds_global_clusters" "test" {
  depends_on = [
    aws_rds_global_cluster.test1,
    aws_rds_global_cluster.test2,
  ]
}
`, rName1, rName2)
}

func testAccGlobalClustersDataSourceConfig_filter(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "postgresql" {
  engine = "aurora-postgresql"
}

data "aws_rds_engine_version" "mysql" {
  engine = "aurora-mysql"
}

resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = data.aws_rds_engine_version.postgresql.engine
  engine_version            = data.aws_rds_engine_version.postgresql.version
}

resource "aws_rds_global_cluster" "decoy" {
  global_cluster_identifier = "%[1]s-decoy"
  engine                    = data.aws_rds_engine_version.mysql.engine
  engine_version            = data.aws_rds_engine_version.mysql.version
}

data "aws_rds_global_clusters" "test" {
  filter {
    name   = "engine"
    values = [data.aws_rds_engine_version.postgresql.engine]
  }

  filter {
    name   = "global_cluster_identifier"
    values = [%[1]q]
  }

  depends_on = [
    aws_rds_global_cluster.test,
    aws_rds_global_cluster.decoy,
  ]
}
`, rName)
}
