---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_service_update_actions"
description: |-
  Provides details about an AWS ElastiCache Service Update Actions.
---

# Data Source: aws_elasticache_service_update_actions

Provides details about an AWS ElastiCache Service Update Actions for a given Cache Cluster or Replication Group.

When creating a new Cache Cluster or Replication Group, it takes approximately 10 minutes for Update Actions to be listed.

## Example Usage

### Basic Usage

The following example will list all Update Actions for the Cache Cluster with a service update status of `available`.

```terraform
data "aws_elasticache_service_update_actions" "example" {
  cache_cluster_id = aws_elasticache_cluster.example.cluster_id

  service_update_status = ["available"]
}
```

## Argument Reference

This data source supports the following arguments:

* `cache_cluster_id` - (Optional) ID of Cache Cluster to list updates for. If neither `cache_cluster_id` nor `replication_group_id` are specified, all service update actions will be listed.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `replication_group_id` - (Optional) ID of Replication Group to list updates for. If neither `replication_group_id` nor `cache_cluster_id` are specified, all service update actions will be listed.
* `service_update_status` - (Optional) Service update statuses to include in list. Valid values are `available`, `cancelled`, and `expired`. If no value is specified, service updates in all statuses will be listed.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `update_actions` - Set of Service Update Actions. Each element has the following attributes:
    * `cache_cluster_id` - ID of Cache Cluster this update action applies to.
    * `engine` - Engine this update applies to.
    * `estimated_update_time` - Estimated duration of update.
    * `replication_group_id` - ID of Replication Group this update action applies to.
    * `service_update_name` - Name of the update.
    * `recommended_apply_by_date` - Date the update should be applied by.
    * `release_date` - Date the update was released.
    * `service_update_severity` - Severity of the update. One of `critical`, `important`, `medium`, or `low`.
    * `service_update_status` - Availability of the update. One of `available`, `cancelled`, or `expired`.
    * `service_update_type` - Type of the update.
    * `update_action_status` - Status of the update action.
