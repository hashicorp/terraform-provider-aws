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
* `resource_id` - (Required) The resource type and unique identifier string for the resource associated with the scaling policy. A non-comprehensive list of examples for this value:

Resource Type | Resource ID Format | Example
------------- | ------------------ | -------
DynamoDB Index | `table/TABLE/index/INDEX` | `table/nameOfTheTable/index/nameOfTheIndex`
DynamoDB Table | `table/TABLE` | `table/nameOfTheTable`
EC2 Spot Fleet Request | `spot-fleet-request/ID` | `spot-fleet-request/sfr-73fbd2ce-aa30-494c-8788-1cee4EXAMPLE`
ECS Service | `service/CLUSTER/SERVICE` | `service/default/sample-webapp`
EMR Instance Group | `instancegroup/CLUSTER/GROUP` | `instancegroup/j-2EEZNYKUA1NTV/ig-1791Y4E1L8YI0`

* `role_arn` - (Required) The ARN of the IAM role that allows Application
AutoScaling to modify your scalable target on your behalf.
* `scalable_dimension` - (Required) The scalable dimension of the scalable target. The scalable dimension contains the service namespace, resource type, and scaling property. A non-comprehensive list of examples for this value:

Resource Type | Scalable Dimensions
------------- | ------------------
DynamoDB Index | `dynamodb:index:ReadCapacityUnits`, `dynamodb:index:WriteCapacityUnits`
DynamoDB Table | `dynamodb:table:ReadCapacityUnits`, `dynamodb:table:WriteCapacityUnits`
EC2 Spot Fleet Request | `ec2:spot-fleet-request:TargetCapacity`
ECS Service | `ecs:service:DesiredCount`
EMR Cluster Instance Group | `elasticmapreduce:instancegroup:InstanceCount`

* `service_namespace` - (Required) The AWS service namespace of the scalable target. A non-comprehensive list of examples for this value:
  * `dynamodb`
  * `ec2`
  * `ecs`
  * `elasticmapreduce`
