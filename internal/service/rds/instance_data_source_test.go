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

func TestAccRDSInstanceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_instance.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAddress, resourceName, names.AttrAddress),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAllocatedStorage, resourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAutoMinorVersionUpgrade, resourceName, names.AttrAutoMinorVersionUpgrade),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_instance_arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_instance_class", resourceName, "instance_class"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_name", resourceName, "db_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_subnet_group", resourceName, "db_subnet_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled_cloudwatch_logs_exports.#", resourceName, "enabled_cloudwatch_logs_exports.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEndpoint, resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrHostedZoneID, resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIOPS, resourceName, names.AttrIOPS),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_username", resourceName, names.AttrUsername),
					resource.TestCheckResourceAttrPair(dataSourceName, "max_allocated_storage", resourceName, "max_allocated_storage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az", resourceName, "multi_az"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_type", resourceName, "network_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceID, resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_throughput", resourceName, "storage_throughput"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStorageType, resourceName, names.AttrStorageType),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccRDSInstanceDataSource_ManagedMasterPassword_managed(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_instance.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_managedMasterPassword(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAddress, resourceName, names.AttrAddress),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAllocatedStorage, resourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAutoMinorVersionUpgrade, resourceName, names.AttrAutoMinorVersionUpgrade),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_instance_arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_instance_class", resourceName, "instance_class"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_name", resourceName, "db_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_subnet_group", resourceName, "db_subnet_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEndpoint, resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrHostedZoneID, resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIOPS, resourceName, names.AttrIOPS),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_username", resourceName, names.AttrUsername),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_user_secret.0.kms_key_id", resourceName, "master_user_secret.0.kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_user_secret.0.secret_arn", resourceName, "master_user_secret.0.secret_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_user_secret.0.secret_status", resourceName, "master_user_secret.0.secret_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az", resourceName, "multi_az"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_type", resourceName, "network_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceID, resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_throughput", resourceName, "storage_throughput"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStorageType, resourceName, names.AttrStorageType),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccRDSInstanceDataSource_matchTags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_instance.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_matchTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAddress, resourceName, names.AttrAddress),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAllocatedStorage, resourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAutoMinorVersionUpgrade, resourceName, names.AttrAutoMinorVersionUpgrade),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_instance_arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_instance_class", resourceName, "instance_class"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_name", resourceName, "db_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_subnet_group", resourceName, "db_subnet_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled_cloudwatch_logs_exports.#", resourceName, "enabled_cloudwatch_logs_exports.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEndpoint, resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrHostedZoneID, resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIOPS, resourceName, names.AttrIOPS),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_username", resourceName, names.AttrUsername),
					resource.TestCheckResourceAttrPair(dataSourceName, "max_allocated_storage", resourceName, "max_allocated_storage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az", resourceName, "multi_az"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_type", resourceName, "network_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceID, resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_throughput", resourceName, "storage_throughput"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStorageType, resourceName, names.AttrStorageType),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccInstanceDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 10
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  username                = "tfacctest"
  max_allocated_storage   = 100

  enabled_cloudwatch_logs_exports = [
    "audit",
    "error",
  ]

  tags = {
    Name = %[1]q
  }
}

data "aws_db_instance" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
}
`, rName))
}

func testAccInstanceDataSourceConfig_managedMasterPassword(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage           = 10
  backup_retention_period     = 0
  db_subnet_group_name        = aws_db_subnet_group.test.name
  engine                      = data.aws_rds_engine_version.default.engine
  engine_version              = data.aws_rds_engine_version.default.version
  identifier                  = %[1]q
  instance_class              = data.aws_rds_orderable_db_instance.test.instance_class
  manage_master_user_password = true
  db_name                     = "test"
  skip_final_snapshot         = true
  username                    = "tfacctest"

  tags = {
    Name = %[1]q
  }
}

data "aws_db_instance" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
}
`, rName))
}

func testAccInstanceDataSourceConfig_matchTags(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		testAccInstanceConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_db_instance" "test" {
  allocated_storage       = 10
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_name                 = "test"
  password                = "avoid-plaintext-passwords"
  skip_final_snapshot     = true
  username                = "tfacctest"
  max_allocated_storage   = 100

  enabled_cloudwatch_logs_exports = [
    "audit",
    "error",
  ]

  tags = {
    Name = %[1]q
  }
}

data "aws_db_instance" "test" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_db_instance.test]
}
`, rName))
}
