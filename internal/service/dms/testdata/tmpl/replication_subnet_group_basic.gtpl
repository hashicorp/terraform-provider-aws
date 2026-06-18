resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = var.rName
  replication_subnet_group_description = "testing"
  subnet_ids                           = aws_subnet.test[*].id
{{- template "tags" . }}
}

{{ template "acctest.ConfigVPCWithSubnets" 3 }}
