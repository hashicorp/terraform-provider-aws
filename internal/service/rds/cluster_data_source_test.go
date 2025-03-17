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

func TestAccRDSClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "backtrack_window", resourceName, "backtrack_window"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrClusterIdentifier, resourceName, names.AttrClusterIdentifier),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDatabaseName, resourceName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_parameter_group_name", resourceName, "db_cluster_parameter_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_subnet_group_name", resourceName, "db_subnet_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_system_id", resourceName, "db_system_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_mode", resourceName, "engine_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngineVersion, resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrHostedZoneID, resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_username", resourceName, "master_username"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_type", resourceName, "network_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func TestAccRDSClusterDataSource_ManagedMasterPassword_managed(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_cluster.test"
	resourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_managedMasterPassword(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "backtrack_window", resourceName, "backtrack_window"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrClusterIdentifier, resourceName, names.AttrClusterIdentifier),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDatabaseName, resourceName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_parameter_group_name", resourceName, "db_cluster_parameter_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_subnet_group_name", resourceName, "db_subnet_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_mode", resourceName, "engine_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngineVersion, resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrHostedZoneID, resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_username", resourceName, "master_username"),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_user_secret.0.kms_key_id", resourceName, "master_user_secret.0.kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_user_secret.0.secret_arn", resourceName, "master_user_secret.0.secret_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_user_secret.0.secret_status", resourceName, "master_user_secret.0.secret_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_type", resourceName, "network_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccClusterDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = %[1]q
  database_name        = "test"
  engine               = "aurora-mysql"
  master_username      = "tfacctest"
  master_password      = "avoid-plaintext-passwords"
  skip_final_snapshot  = true
  db_subnet_group_name = aws_db_subnet_group.test.name

  tags = {
    Name = %[1]q
  }
}

data "aws_rds_cluster" "test" {
  cluster_identifier = aws_rds_cluster.test.cluster_identifier
}
`, rName))
}

func testAccClusterDataSourceConfig_managedMasterPassword(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_rds_cluster" "test" {
  cluster_identifier          = %[1]q
  database_name               = "test"
  engine                      = "aurora-mysql"
  db_subnet_group_name        = aws_db_subnet_group.test.name
  manage_master_user_password = true
  master_username             = "tfacctest"
  skip_final_snapshot         = true

  tags = {
    Name = %[1]q
  }
}

data "aws_rds_cluster" "test" {
  cluster_identifier = aws_rds_cluster.test.cluster_identifier
}
`, rName))
}
