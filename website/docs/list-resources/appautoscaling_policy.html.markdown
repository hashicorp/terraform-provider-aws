---
subcategory: "Application Auto Scaling"
layout: "aws"
page_title: "AWS: aws_appautoscaling_policy"
description: |-
  Lists Application Auto Scaling Policy resources.
---

# List Resource: aws_appautoscaling_policy

Lists Application Auto Scaling Policy resources.

## Example Usage

```terraform
list "aws_appautoscaling_policy" "example" {
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
