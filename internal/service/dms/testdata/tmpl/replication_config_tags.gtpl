resource "aws_dms_replication_config" "test" {
{{- template "region" }}
  replication_config_identifier = var.rName
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
{{- template "tags" . }}
}

# testAccReplicationConfigConfig_base_DummyDatabase

resource "aws_dms_replication_subnet_group" "test" {
{{- template "region" }}
  replication_subnet_group_id          = var.rName
  replication_subnet_group_description = "terraform test"
  subnet_ids                           = aws_subnet.test[*].id
}

# testAccReplicationEndpointConfig_dummyDatabase

data "aws_partition" "current" {}
data "aws_region" "current" {
{{- template "region" -}}
}

resource "aws_dms_endpoint" "source" {
{{- template "region" }}
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
{{- template "region" }}
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
