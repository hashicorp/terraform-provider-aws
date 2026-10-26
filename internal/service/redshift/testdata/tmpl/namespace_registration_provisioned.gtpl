resource "aws_redshift_namespace_registration" "test" {
{{- template "region" }}
  consumer_identifier            = format("DataCatalog/%s", data.aws_caller_identity.current.account_id)
  namespace_type                 = "provisioned"
  provisioned_cluster_identifier = aws_redshift_cluster.test.cluster_identifier
}

resource "aws_redshift_cluster" "test" {
{{- template "region" }}
  cluster_identifier  = var.rName
  database_name       = "test"
  master_username     = "testuser"
  master_password     = "Testpass123"
  node_type           = "ra3.large"
  cluster_type        = "single-node"
  skip_final_snapshot = true
}

data "aws_caller_identity" "current" {}
