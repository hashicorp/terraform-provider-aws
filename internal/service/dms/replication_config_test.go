// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"fmt"
	"testing"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDMSServerlessReplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessReplication_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_config_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication", "resource_identifier"},
			},
		},
	})
}

func TestAccDMSServerlessReplication_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessReplication_update(rName, "cdc", 2, 16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_config_arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", "2"),
				),
			},
			{
				Config: testAccServerlessReplication_update(rName, "cdc", 4, 32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_config_arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "32"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", "4"),
				),
			},
		},
	})
}

func TestAccDMSServerlessReplication_startReplication(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessReplication_startReplication(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "start_replication", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication", "resource_identifier"},
			},
			{
				Config: testAccServerlessReplication_startReplication(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "start_replication", "false"),
				),
			},
		},
	})
}

func testAccCheckReplicationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn(ctx)

		_, err := tfdms.FindReplicationById(ctx, rs.Primary.ID, conn)

		return err
	}
}

func testAccCheckReplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_replication_config" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn(ctx)

			_, err := tfdms.FindReplicationById(ctx, rs.Primary.ID, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS serverless replication config (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccServerlessReplication_basic(rName string) string {
	return acctest.ConfigCompose(
		serverlessReplicationDBToDB(rName),
		serverlessReplicationBase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  resource_identifier           = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }
}
`, rName))
}

func testAccServerlessReplication_update(rName, replicationType string, minCapacity, maxCapacity int) string {
	return acctest.ConfigCompose(
		serverlessReplicationDBToDB(rName),
		serverlessReplicationBase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  resource_identifier           = %[1]q
  replication_type              = %[2]q
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "%[3]d"
    min_capacity_units           = "%[4]d"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }
}
`, rName, replicationType, maxCapacity, minCapacity))
}

func testAccServerlessReplication_startReplication(rName string, start bool) string {
	return acctest.ConfigCompose(
		serverlessReplicationDBToDB(rName),
		serverlessReplicationBase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  resource_identifier           = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  start_replication = %[2]t

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }
}
`, rName, start))
}

func serverlessReplicationDBToDB(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name        = "%[1]s-pg-cluster"
  family      = "aurora-mysql5.7"
  description = "DMS cluster parameter group"

  parameter {
    name         = "binlog_format"
    value        = "ROW"
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "binlog_row_image"
    value        = "Full"
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "binlog_checksum"
    value        = "NONE"
    apply_method = "pending-reboot"
  }
}

resource "aws_rds_cluster" "test1" {
  cluster_identifier              = "%[1]s-aurora-cluster-source"
  engine                          = "aurora-mysql"
  engine_version                  = "5.7.mysql_aurora.2.11.2"
  database_name                   = "tftest"
  master_username                 = "tftest"
  master_password                 = "mustbeeightcharaters"
  skip_final_snapshot             = true
  vpc_security_group_ids          = [aws_security_group.test.id]
  db_subnet_group_name            = aws_db_subnet_group.test.name
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.test.name
}

resource "aws_rds_cluster_instance" "test1" {
  identifier           = "%[1]s-test1-primary"
  cluster_identifier   = aws_rds_cluster.test1.id
  instance_class       = "db.t2.small"
  engine               = aws_rds_cluster.test1.engine
  engine_version       = aws_rds_cluster.test1.engine_version
  db_subnet_group_name = aws_db_subnet_group.test.name
}

resource "aws_rds_cluster" "test2" {
  cluster_identifier     = "%[1]s-aurora-cluster-target"
  engine                 = "aurora-mysql"
  engine_version         = "5.7.mysql_aurora.2.11.2"
  database_name          = "tftest"
  master_username        = "tftest"
  master_password        = "mustbeeightcharaters"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
  db_subnet_group_name   = aws_db_subnet_group.test.name
}

resource "aws_rds_cluster_instance" "test2" {
  identifier           = "%[1]s-test2-primary"
  cluster_identifier   = aws_rds_cluster.test2.id
  instance_class       = "db.t2.small"
  engine               = aws_rds_cluster.test2.engine
  engine_version       = aws_rds_cluster.test2.engine_version
  db_subnet_group_name = aws_db_subnet_group.test.name
}

resource "aws_dms_endpoint" "target" {
  database_name = "tftest"
  endpoint_id   = "%[1]s-target"
  endpoint_type = "target"
  engine_name   = "aurora"
  server_name   = aws_rds_cluster.test2.endpoint
  port          = 3306
  username      = "tftest"
  password      = "mustbeeightcharaters"
}

resource "aws_dms_endpoint" "source" {
  database_name = "tftest"
  endpoint_id   = "%[1]s-source"
  endpoint_type = "source"
  engine_name   = "aurora"
  server_name   = aws_rds_cluster.test1.endpoint
  port          = 3306
  username      = "tftest"
  password      = "mustbeeightcharaters"
}
`, rName))
}

func serverlessReplicationBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "%[1]s-vpc"
  }
}

resource "aws_subnet" "test1" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-subnet-test1"
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-subnet-test2"
  }
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = [aws_subnet.test1.id, aws_subnet.test2.id]
}

resource "aws_db_subnet_group" "test" {
  name = %[1]q
  subnet_ids = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
  ]
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = -1
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName))
}
