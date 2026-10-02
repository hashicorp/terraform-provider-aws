---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_apply_service_update"
description: |-
  Applies ElastiCache service updates to a specified cache cluster or replication group.
---

# Action: aws_elasticache_apply_service_update

Applies ElastiCache service updates to a specified cache cluster or replication group.

For information about applying service updates, see the [Service updates in ElastiCache](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/Self-Service-Updates.html) in the [Amazon ElastiCache User Guide](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/WhatIs.html).

## Example Usage

### Basic Usage

```terraform
action "aws_elasticache_apply_service_update" "example" {
  config {
    replication_group_id = aws_elasticache_replication_group.replication_group_id
    service_update_name  = "example-service-update"
  }
}

resource "aws_elasticache_replication_group" "example" {
  # ... replication group configuration
}

resource "terraform_data" "example" {
  input = "trigger"

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_elasticache_apply_service_update.example]
    }
  }
}
```

## Argument Reference

This action supports the following arguments:

* `cache_cluster_id` - (Optional) ID of Cache Cluster to apply update to. One of `cache_cluster_id` or `replication_group_id` is required.
* `region` - (Optional) Region where this action will be [invoked](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `replication_group_id` - (Optional) ID of Replication Group to apply update to. One of `replication_group_id` or `cache_cluster_id` is required.
* `service_update_name` - (Required) Name of service update to apply.

## Timeouts

Configuration options:

* `invoke` - (Default `60m`)
