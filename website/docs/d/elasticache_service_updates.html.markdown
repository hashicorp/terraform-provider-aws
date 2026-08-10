---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_service_updates"
description: |-
  Provides details about AWS ElastiCache Service Updates.
---

# Data Source: aws_elasticache_service_updates

Provides details about AWS ElastiCache Service Updates.

## Example Usage

### Basic Usage

```terraform
data "aws_elasticache_service_updates" "example" {
  status = ["available"]
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `status` - (Optional) Set of one or more Service Update statuses. Elements must be one of `available`, `cancelled`, or `expired`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `service_updates` - Set of Service Updates. Each element has the following attributes:
    * `auto_update_after_recommended_apply_by_date` - Whether the update will be applied after `recommended_apply_by_date`.
    * `engine` - Engine this update applies to.
    * `engine_version` - Engine version this update applies to.
    * `estimated_update_time` - Estimated duration of update.
    * `description` - Description of the update.
    * `end_date` - Date the update will no longer be available.
    * `name` - Name of the update.
    * `recommended_apply_by_date` - Date the update should be applied by.
    * `release_date` - Date the update was released.
    * `severity` - Severity of the update. One of `critical`, `important`, `medium`, or `low`.
    * `status` - Availability of the update. One of `available`, `cancelled`, or `expired`.
    * `type` - Type of the update.
