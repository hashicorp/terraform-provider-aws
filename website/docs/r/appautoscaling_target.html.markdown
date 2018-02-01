---
layout: "aws"
page_title: "AWS: aws_appautoscaling_target"
sidebar_current: "docs-aws-resource-appautoscaling-target"
description: |-
  Provides an Application AutoScaling ScalableTarget resource.
---

# aws_appautoscaling_target

Provides an Application AutoScaling ScalableTarget resource.

## Example Usage

### DynamoDB Table Autoscaling

```hcl
resource "aws_appautoscaling_target" "dynamodb_table_read_target" {
  max_capacity       = 100
  min_capacity       = 5
  resource_id        = "table/tableName"
  role_arn           = "${data.aws_iam_role.DynamoDBAutoscaleRole.arn}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}
```

### DynamoDB Index Autoscaling

```hcl
resource "aws_appautoscaling_target" "dynamodb_index_read_target" {
  max_capacity       = 100
  min_capacity       = 5
  resource_id        = "table/tableName/index/indexName"
  role_arn           = "${data.aws_iam_role.DynamoDBAutoscaleRole.arn}"
  scalable_dimension = "dynamodb:index:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}
```

### ECS Service Autoscaling

```hcl
resource "aws_appautoscaling_target" "ecs_target" {
  max_capacity       = 4
  min_capacity       = 1
  resource_id        = "service/clusterName/serviceName"
  role_arn           = "${var.ecs_iam_role}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}
```

## Argument Reference

The following arguments are supported:

* `max_capacity` - (Required) The max capacity of the scalable target.
* `min_capacity` - (Required) The min capacity of the scalable target.
* `resource_id` - (Required) The resource type and unique identifier string for the resource associated with the scaling policy. Documentation can be found in the `ResourceId` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html#API_RegisterScalableTarget_RequestParameters)
* `role_arn` - (Optional) The ARN of the IAM role that allows Application
AutoScaling to modify your scalable target on your behalf.
* `scalable_dimension` - (Required) The scalable dimension of the scalable target. Documentation can be found in the `ScalableDimension` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html#API_RegisterScalableTarget_RequestParameters)
* `service_namespace` - (Required) The AWS service namespace of the scalable target. Documentation can be found in the `ServiceNamespace` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html#API_RegisterScalableTarget_RequestParameters)
