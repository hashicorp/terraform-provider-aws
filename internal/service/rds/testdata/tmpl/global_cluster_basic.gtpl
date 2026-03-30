resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = var.rName
  engine                    = "aurora-postgresql"

{{- template "tags" . }}
}