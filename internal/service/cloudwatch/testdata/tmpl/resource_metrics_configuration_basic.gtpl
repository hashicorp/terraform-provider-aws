resource "aws_cloudwatch_resource_metrics_configuration" "test" {
{{- template "region" }}
  resource_arn = aws_elasticache_replication_group.test.arn
}

resource "aws_elasticache_replication_group" "test" {
{{- template "region" }}
  replication_group_id = var.rName
  description          = "terraform-provider-aws acceptance testing"
  engine               = "valkey"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1
  apply_immediately    = true
}
