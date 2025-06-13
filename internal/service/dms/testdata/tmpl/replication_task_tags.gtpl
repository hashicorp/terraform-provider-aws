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
{{- template "tags" . }}
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

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
