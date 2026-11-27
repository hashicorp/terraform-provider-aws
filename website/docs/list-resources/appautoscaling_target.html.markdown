---
subcategory: "Application Auto Scaling"
layout: "aws"
page_title: "AWS: aws_appautoscaling_target"
description: |-
  Lists Application Auto Scaling scalable targets.
---

# List Resource: aws_appautoscaling_target

Lists Application Auto Scaling scalable targets in the specified service namespace.

## Example Usage

```terraform
list "aws_appautoscaling_target" "example" {
  provider = aws

  config {
    service_namespace = "ecs"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
* `service_namespace` - (Required) Namespace of the AWS service that owns the scalable target. Valid values: `appstream`, `cassandra`, `comprehend`, `custom-resource`, `dynamodb`, `ec2`, `ecs`, `elasticache`, `elasticmapreduce`, `kafka`, `lambda`, `neptune`, `rds`, `sagemaker`, `workspaces`.
