# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_dms_replication_task" "test" {
  replication_task_id      = var.rName
  migration_type           = "full-load"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%",
            "table-name"  = "%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

# testAccReplicationTaskConfig_base

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = var.rName
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_replication_instance" "test" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.t3.medium"
  replication_instance_id      = var.rName
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}

# testAccReplicationEndpointConfig_dummyDatabase

data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_dms_endpoint" "source" {
  database_name = var.rName
  endpoint_id   = "${var.rName}-source"
  endpoint_type = "source"
  engine_name   = "aurora"
  server_name   = "tf-test-cluster.cluster-xxxxxxx.${data.aws_region.current.name}.rds.${data.aws_partition.current.dns_suffix}"
  port          = 3306
  username      = "tftest"
  password      = "tftest"
}

resource "aws_dms_endpoint" "target" {
  database_name = var.rName
  endpoint_id   = "${var.rName}-target"
  endpoint_type = "target"
  engine_name   = "aurora"
  server_name   = "tf-test-cluster.cluster-xxxxxxx.${data.aws_region.current.name}.rds.${data.aws_partition.current.dns_suffix}"
  port          = 3306
  username      = "tftest"
  password      = "tftest"
}

# acctest.ConfigVPCWithSubnets(rName, 2)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
