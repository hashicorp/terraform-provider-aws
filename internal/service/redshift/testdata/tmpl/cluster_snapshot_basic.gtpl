resource "aws_redshift_cluster_snapshot" "test" {
{{- template "region" }}
  cluster_identifier  = aws_redshift_cluster.test.cluster_identifier
  snapshot_identifier = var.rName
{{- template "tags" . }}
}

# testAccClusterConfig_basic

resource "aws_redshift_cluster" "test" {
{{- template "region" }}
  cluster_identifier    = var.rName
  database_name         = "mydb"
  master_username       = "foo_test"
  master_password       = "Mustbe8characters"
  node_type             = "ra3.large"
  allow_version_upgrade = false
  skip_final_snapshot   = true
}
