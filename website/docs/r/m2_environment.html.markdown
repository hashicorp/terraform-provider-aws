---
subcategory: "m2"
layout: "aws"
page_title: "AWS: aws_m2_environment"
description: |-
  Provides a resource to create a M2 environment in a AWS Account.
---

# Resource: aws_m2_environment

Provides a resource to createe an [AWS Mainframe Modernization](https://docs.aws.amazon.com/m2/latest/APIReference/API_CreateEnvironment.html).

## Example Usage

```terraform
resource "aws_m2_environment" "test" {
  engine_type   = "microfocus"
  instance_type = "M2.m5.large"
  name          = "Microfocus M2 Environment
}
```

## Argument Reference

This resource supports the following arguments:

* `client_token` -  (Auto Generaated) Unique, case-sensitive identifier you provide to ensure the idempotency of the request to create an environment. 
* `description` - (Optional) The description of the runtime environment.
* `engine_type` - (Required) The engine type for the runtime environment. Valid Values are `microfocus` or `blueage`
* `engine_version` - (Opional) The version of the engine type for the runtime environment.
* `high_availability_config` - (Optional) The details of a high availability configuration for this runtime environment. Check the high availability config object for more details.
* `instance_type` - (Required) The type of instance for the runtime environment.
* `kms_key_id` - (Optional) The identifier of a customer managed key.
* `name` - (Required) The name of the runtime environment. Must be unique within the account.
* `preferred_maintainence_window` - (Optional) Configures the maintenance window that you want for the runtime environment. The maintenance window must have the format ddd:hh24:mi-ddd:hh24:mi and must be less than 24 hours. The following two examples are valid maintenance windows: sun:23:45-mon:00:15 or sat:01:00-sat:03:00. If you do not provide a value, a random system-generated value will be assigned.
* `publicly_accessible` - (Optional) Flag which specifies whether the runtime environment is publicly accessible.
* `security_group_ids` - (Optional) The list of security groups for the VPC associated with this runtime environment.
* `storage_configurations` - (Optional) The array of storage configurations for this runtime environment. Check the storage configurations object for more details.
* `subnet_ids` - (Optional) The list of subnets associated with the VPC for this runtime environment.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `high_availability_config` object supports the following:

* `desired_capacity` -  (Required) The number of instances in a high availability configuration. The minimum possible value is 1 and the maximum is 100.

The `storage_configurations` object supports the following and only one configuration can be specified. 

* `efs` -  (Optional) Defines the storage configuration for an Amazon EFS file system.
* `fsx` -  (Optional) Defines the storage configuration for an Amazon FSx file system.

The `efs` or `fsx`  object supports the following fields.

* `file_system_id` -  (Required) The file system identifier.
* `mount_point` -  (Required) The mount point for the file system.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `environment_id` - The unique identifier of the runtime environment.


## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_m2_environment` using the environmentid. For example:

```terraform
import {
  to = aws_m2_environment.example
  environment_id = "abcd1"
}
```

Using `terraform import`, import `aws_m2_environment` using the environmentid. For example:

```console
% terraform import aws_m2_environment.example abcd1
```
