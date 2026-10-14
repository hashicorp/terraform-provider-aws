# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_resource_metrics_configuration" "test" {
  region = var.region

  resource_arn = aws_elasticache_replication_group.test.arn
}

resource "aws_elasticache_replication_group" "test" {
  region = var.region

  replication_group_id = var.rName
  description          = "terraform-provider-aws acceptance testing"
  engine               = "valkey"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1
  apply_immediately    = true
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
