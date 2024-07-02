// Copyright (c) HashiCorp, Inc.
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

func TestAccRDSGlobalClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_global_cluster.test"
	resourceName := "aws_rds_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.ConfigMultipleRegionProvider(2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "database_name", resourceName, "database_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "deletion_protection", resourceName, "deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_identifier", resourceName, "global_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_members", resourceName, "global_cluster_members"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_resource_id", resourceName, "global_cluster_resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_encrypted", resourceName, "storage_encrypted"),
				),
			},
		},
	})
}

func testAccGlobalClusterDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`

data "aws_region" "alternate" {
	provider = "awsalternate"
}
		
data "aws_region" "current" {}
resource "aws_rds_global_cluster" "test" {
		global_cluster_identifier = "%[1]s"
		engine                    = "aurora-postgresql"
		engine_version            = "15.5"
		database_name             = "example_db"
}

resource "aws_rds_cluster" "primary" {
		engine                    = aws_rds_global_cluster.test.engine
		engine_version            = aws_rds_global_cluster.test.engine_version
		cluster_identifier        = "test-primary-cluster"
		master_username           = "username"
		master_password           = "somepass123"
		database_name             = "example_db"
		global_cluster_identifier = aws_rds_global_cluster.test.id
		skip_final_snapshot       = true
		
		depends_on = [
			aws_rds_global_cluster.test
		]
}

resource "aws_rds_cluster_instance" "primary" {
		identifier           = "test-primary-cluster-instance"
		cluster_identifier   = aws_rds_cluster.primary.id
		instance_class       = "db.r5.large"
		engine               = aws_rds_global_cluster.test.engine
		engine_version       = aws_rds_global_cluster.test.engine_version
		
		depends_on = [
			aws_rds_cluster.primary
		]
}

resource "aws_rds_cluster" "secondary" {
		provider                  = "awsalternate"
		engine                    = aws_rds_global_cluster.test.engine
		engine_version            = aws_rds_global_cluster.test.engine_version
		cluster_identifier        = "test-secondary-cluster"
		global_cluster_identifier = aws_rds_global_cluster.test.id
		replication_source_identifier = aws_rds_cluster.primary.arn
		source_region 			  = data.aws_region.alternate.name
		skip_final_snapshot       = true
		depends_on = [
			aws_rds_cluster_instance.primary
		]
}

resource "aws_rds_cluster_instance" "secondary" {
		provider             = "awsalternate"
		identifier           = "test-secondary-cluster-instance"
		cluster_identifier   = aws_rds_cluster.secondary.id
		instance_class       = "db.r5.large"
		engine               = aws_rds_global_cluster.test.engine
		engine_version       = aws_rds_global_cluster.test.engine_version
	  
		depends_on = [
			aws_rds_cluster.secondary
		]
}

data "aws_rds_global_cluster" "test" {
	global_cluster_identifier = aws_rds_global_cluster.test.global_cluster_identifier
}
`, rName))
}
