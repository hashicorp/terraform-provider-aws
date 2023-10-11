---
subcategory: "Application Auto Scaling"
layout: "aws"
page_title: "AWS: aws_appautoscaling_target"
description: |-
  Provides an Application AutoScaling ScalableTarget resource.
---

# Resource: aws_appautoscaling_target

Provides an Application AutoScaling ScalableTarget resource. To manage policies which get attached to the target, see the [`aws_appautoscaling_policy` resource](/docs/providers/aws/r/appautoscaling_policy.html).

~> **NOTE:** Scalable targets created before 2023-03-20 may not have an assigned `arn`. These resource cannot use `tags` or participate in `default_tags`. To prevent `terraform plan` showing differences that can never be reconciled, use the [`lifecycle.ignore_changes`](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes) meta-argument. See the example below.

~> **NOTE:** The [Application Auto Scaling service automatically attempts to manage IAM Service-Linked Roles](https://docs.aws.amazon.com/autoscaling/application/userguide/security_iam_service-with-iam.html#security_iam_service-with-iam-roles) when registering certain service namespaces for the first time. To manually manage this role, see the [`aws_iam_service_linked_role` resource](/docs/providers/aws/r/iam_service_linked_role.html).

## Example Usage

### DynamoDB Table Autoscaling

```terraform
resource "aws_appautoscaling_target" "dynamodb_table_read_target" {
  max_capacity       = 100
  min_capacity       = 5
  resource_id        = "table/${aws_dynamodb_table.example.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}
```

### DynamoDB Index Autoscaling

```terraform
resource "aws_appautoscaling_target" "dynamodb_index_read_target" {
  max_capacity       = 100
  min_capacity       = 5
  resource_id        = "table/${aws_dynamodb_table.example.name}/index/${var.index_name}"
  scalable_dimension = "dynamodb:index:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}
```

### ECS Service Autoscaling

```terraform
resource "aws_appautoscaling_target" "ecs_target" {
  max_capacity       = 4
  min_capacity       = 1
  resource_id        = "service/${aws_ecs_cluster.example.name}/${aws_ecs_service.example.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}
```

### Aurora Read Replica Autoscaling

```terraform
resource "aws_appautoscaling_target" "replicas" {
  service_namespace  = "rds"
  scalable_dimension = "rds:cluster:ReadReplicaCount"
  resource_id        = "cluster:${aws_rds_cluster.example.id}"
  min_capacity       = 1
  max_capacity       = 15
}
```

### Suppressing `tags_all` Differences For Older Resources

```terraform
resource "aws_appautoscaling_target" "ecs_target" {
  max_capacity       = 4
  min_capacity       = 1
  resource_id        = "service/${aws_ecs_cluster.example.name}/${aws_ecs_service.example.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"

  lifecycle {
    ignore_changes = [
      tags_all,
    ]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `max_capacity` - (Required) Max capacity of the scalable target.
* `min_capacity` - (Required) Min capacity of the scalable target.
* `resource_id` - (Required) Resource type and unique identifier string for the resource associated with the scaling policy. Documentation can be found in the `ResourceId` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html#API_RegisterScalableTarget_RequestParameters)
* `role_arn` - (Optional) ARN of the IAM role that allows Application AutoScaling to modify your scalable target on your behalf. This defaults to an IAM Service-Linked Role for most services and custom IAM Roles are ignored by the API for those namespaces. See the [AWS Application Auto Scaling documentation](https://docs.aws.amazon.com/autoscaling/application/userguide/security_iam_service-with-iam.html#security_iam_service-with-iam-roles) for more information about how this service interacts with IAM.
* `scalable_dimension` - (Required) Scalable dimension of the scalable target. Documentation can be found in the `ScalableDimension` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html#API_RegisterScalableTarget_RequestParameters)
* `service_namespace` - (Required) AWS service namespace of the scalable target. Documentation can be found in the `ServiceNamespace` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html#API_RegisterScalableTarget_RequestParameters)
* `tags` - (Optional) Map of tags to assign to the scalable target. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the scalable target.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Application AutoScaling Target using the `service-namespace` , `resource-id` and `scalable-dimension` separated by `/`. For example:

```terraform
import {
  to = aws_appautoscaling_target.test-target
  id = "service-namespace/resource-id/scalable-dimension"
}
```

Using `terraform import`, import Application AutoScaling Target using the `service-namespace` , `resource-id` and `scalable-dimension` separated by `/`. For example:

```console
% terraform import aws_appautoscaling_target.test-target service-namespace/resource-id/scalable-dimension
```
