// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSGlobalClustersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_global_clusters.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClustersDataSourceConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "global_cluster_identifiers.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "global_cluster_arns.#"),
				),
			},
		},
	})
}

func TestAccRDSGlobalClustersDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_global_clusters.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClustersDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "global_cluster_identifiers.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_identifiers.0", resourceName, "global_cluster_identifier"),
				),
			},
		},
	})
}

func testAccGlobalClustersDataSourceConfig_basic(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test1" {
  global_cluster_identifier = %[1]q
  engine                    = "aurora-postgresql"
  engine_version            = "16.4"
}

resource "aws_rds_global_cluster" "test2" {
  global_cluster_identifier = %[2]q
  engine                    = "aurora-postgresql"
  engine_version            = "16.4"
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
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "aurora-postgresql"
  engine_version            = "16.4"
}

resource "aws_rds_global_cluster" "decoy" {
  global_cluster_identifier = "%[1]s-decoy"
  engine                    = "aurora-mysql"
  engine_version            = "8.0.mysql_aurora.3.07.1"
}

data "aws_rds_global_clusters" "test" {
  filter {
    name   = "engine"
    values = ["aurora-postgresql"]
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
