resource "aws_redshift_usage_limit" "test" {
{{- template "region" }}
  cluster_identifier = aws_redshift_cluster.test.id
  feature_type       = "concurrency-scaling"
  limit_type         = "time"
  amount             = 60
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
