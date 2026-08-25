resource "aws_rds_cluster" "test" {
{{- template "region" }}
  cluster_identifier  = var.rName
  database_name       = "test"
  engine              = "aurora-mysql"
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
{{- template "tags" . }}
}
