---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_cluster"
description: |-
    Provides details about an ecs cluster
---

# Data Source: aws_ecs_cluster

The ECS Cluster data source allows access to details of a specific
cluster within an AWS ECS service.

## Example Usage

```terraform
data "aws_ecs_cluster" "ecs-mongo" {
  cluster_name = "ecs-mongo-production"
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster_name` - (Required) Name of the ECS Cluster
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `active_services_count` - Number of services running on the ECS Cluster in an `ACTIVE` state
* `arn` - ARN of the ECS Cluster
* `attachments` - Resources attached to the ECS Cluster, such as Auto Scaling group capacity provider policies. See [`attachments`](#attachments) below.
* `attachments_status` - Status of the capacity providers associated with the ECS Cluster
* `capacity_providers` - Capacity providers associated with the ECS Cluster
* `configuration` - Execute command and managed storage configuration for the ECS Cluster. See [`configuration`](#configuration) below.
* `default_capacity_provider_strategy` - Default capacity provider strategy for the ECS Cluster. See [`default_capacity_provider_strategy`](#default_capacity_provider_strategy) below.
* `pending_tasks_count` - Number of pending tasks for the ECS Cluster
* `registered_container_instances_count` - Number of registered container instances for the ECS Cluster
* `running_tasks_count` - Number of running tasks for the ECS Cluster
* `service_connect_defaults` - Default Service Connect namespace
* `setting` - Settings associated with the ECS Cluster
* `statistics` - Additional statistics about the ECS Cluster, such as running and pending task counts by launch type. Each element has a `name` and a `value`.
* `status` - Status of the ECS Cluster
* `tags` - Key-value map of resource tags

### `attachments`

* `details` - Details of the attachment. Each element has a `name` and a `value`.
* `id` - Unique identifier of the attachment
* `status` - Status of the attachment
* `type` - Type of the attachment, such as `as_policy`

### `configuration`

* `execute_command_configuration` - Details of the execute command configuration. See [`execute_command_configuration`](#execute_command_configuration) below.
* `managed_storage_configuration` - Details of the managed storage configuration. See [`managed_storage_configuration`](#managed_storage_configuration) below.

### `execute_command_configuration`

* `kms_key_id` - AWS Key Management Service key ID used to encrypt the data between the local client and the container
* `log_configuration` - Log configuration for the results of the execute command actions. See [`log_configuration`](#log_configuration) below.
* `logging` - Log setting to use for redirecting logs for execute command results. One of `NONE`, `DEFAULT` or `OVERRIDE`.

### `log_configuration`

* `cloud_watch_encryption_enabled` - Whether encryption for CloudWatch logs is enabled
* `cloud_watch_log_group_name` - Name of the CloudWatch log group to send logs to
* `s3_bucket_encryption_enabled` - Whether encryption for S3 logs is enabled
* `s3_bucket_name` - Name of the S3 bucket to send logs to
* `s3_key_prefix` - Optional folder in the S3 bucket to place logs in

### `managed_storage_configuration`

* `fargate_ephemeral_storage_kms_key_id` - AWS Key Management Service key ID for the Fargate ephemeral storage
* `kms_key_id` - AWS Key Management Service key ID used to encrypt the managed storage

### `default_capacity_provider_strategy`

* `base` - Minimum number of tasks to run on the specified capacity provider
* `capacity_provider` - Name of the capacity provider
* `weight` - Relative percentage of the total number of launched tasks that should use the specified capacity provider
