// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSInstancesDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_instances.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(dataSourceName, "instance_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_arns.0", resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "instance_identifiers.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_identifiers.0", resourceName, names.AttrIdentifier),
				),
			},
		},
	})
}

func TestAccRDSInstancesDataSource_matchTags(t *testing.T) {
	ctx := acctest.Context(t)
	var dbInstance rds.DBInstance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_instances.test"
	resourceName := "aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_matchTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists(ctx, resourceName, &dbInstance),
					resource.TestCheckResourceAttr(dataSourceName, "instance_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_arns.0", resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "instance_identifiers.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_identifiers.0", resourceName, names.AttrIdentifier),
				),
			},
		},
	})
}

func testAccInstancesDataSourceConfig_filter(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "postgres"
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  allocated_storage    = 10
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = "db.t4g.micro"
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true

  apply_immediately = true
}

resource "aws_db_instance" "wrong" {
  identifier           = "%[1]s-wrong"
  allocated_storage    = 10
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = "db.t4g.micro"
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true

  apply_immediately = true
}


data "aws_db_instances" "test" {
  filter {
    name   = "db-instance-id"
    values = [aws_db_instance.test.identifier]
  }

  depends_on = [aws_db_instance.wrong]
}
`, rName)
}

func testAccInstancesDataSourceConfig_matchTags(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "postgres"
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  allocated_storage    = 10
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = "db.t4g.micro"
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true

  apply_immediately = true

  tags = {
    Name = %[1]q
    Test = "true"
  }
}

resource "aws_db_instance" "wrong" {
  identifier           = "%[1]s-wrong"
  allocated_storage    = 10
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = "db.t4g.micro"
  db_name              = "test"
  password             = "avoid-plaintext-passwords"
  username             = "tfacctest"
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true

  apply_immediately = true

  tags = {
    Name = "%[1]s-wrong"
    Test = "true"
  }
}

data "aws_db_instances" "test" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_db_instance.test, aws_db_instance.wrong]
}
`, rName)
}
