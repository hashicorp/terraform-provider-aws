// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rdsdata_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSDataQueryDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rdsdata_query.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "records"),
					resource.TestCheckResourceAttr(dataSourceName, "sql", "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA LIMIT 1"),
				),
			},
		},
	})
}

func TestAccRDSDataQueryDataSource_withParameters(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rdsdata_query.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDataSourceConfig_withParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "records"),
					resource.TestCheckResourceAttr(dataSourceName, "sql", "SELECT :param1 as test_column"),
					resource.TestCheckResourceAttr(dataSourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "parameters.0.name", "param1"),
					resource.TestCheckResourceAttr(dataSourceName, "parameters.0.value", "test_value"),
				),
			},
		},
	})
}

func testAccQueryDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccQueryDataSourceConfig_base(rName), `
data "aws_rdsdata_query" "test" {
  depends_on   = [aws_rds_cluster_instance.test]
  resource_arn = aws_rds_cluster.test.arn
  secret_arn   = aws_secretsmanager_secret.test.arn
  sql          = "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA LIMIT 1"
}
`)
}

func testAccQueryDataSourceConfig_withParameters(rName string) string {
	return acctest.ConfigCompose(testAccQueryDataSourceConfig_base(rName), `
data "aws_rdsdata_query" "test" {
  depends_on   = [aws_rds_cluster_instance.test]
  resource_arn = aws_rds_cluster.test.arn
  secret_arn   = aws_secretsmanager_secret.test.arn
  sql          = "SELECT :param1 as test_column"

  parameters {
    name  = "param1"
    value = "test_value"
  }
}
`)
}

func testAccQueryDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier           = %[1]q
  engine                       = "aurora-mysql"
  database_name                = "test"
  master_username              = "username"
  master_password              = "mustbeeightcharacters"
  backup_retention_period      = 7
  preferred_backup_window      = "07:00-09:00"
  preferred_maintenance_window = "tue:04:00-tue:04:30"
  skip_final_snapshot          = true
  enable_http_endpoint         = true

  serverlessv2_scaling_configuration {
    max_capacity = 8
    min_capacity = 0.5
  }
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = "db.serverless"
  engine             = aws_rds_cluster.test.engine
  engine_version     = aws_rds_cluster.test.engine_version
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({
    username = aws_rds_cluster.test.master_username
    password = aws_rds_cluster.test.master_password
  })
}
`, rName)
}
