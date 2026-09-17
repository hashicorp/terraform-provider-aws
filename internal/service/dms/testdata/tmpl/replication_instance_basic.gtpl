data "aws_partition" "current" {}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = var.rName
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.id
{{- template "tags" . }}
}

# testAccReplicationInstanceConfig_base

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = var.rName
  replication_subnet_group_description = "testing"
  subnet_ids                           = aws_subnet.test[*].id
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
