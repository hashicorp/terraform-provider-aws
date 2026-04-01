---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_pool"
description: |-
  Terraform data source for managing an AWS WorkSpaces Pool.
---
# Data Source: aws_workspaces_pool

Terraform data source for managing an AWS WorkSpaces Pool.

## Example Usage

### Basic Usage with ID

```terraform
data "aws_workspaces_pool" "example" {
  id = "wspool-12345678"
}
```

### Basic Usage with Name

```terraform
data "aws_workspaces_pool" "example" {
  name = "example-pool"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - ID of the WorkSpaces Pool.
* `name` - Name of the WorkSpaces Pool.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference)

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `application_settings` - Information about the application settings for the WorkSpaces Pool.
    * `s3_bucket_name` - S3 bucket name for the application settings.
    * `settings_group` - Name of the settings group for the application settings.
    * `status` - Status of the application settings.
* `arn` - ARN of the WorkSpaces Pool.
* `bundle_id` - ID of the bundle for the WorkSpaces Pool.
* `capacity` - Information about the capacity of the WorkSpaces Pool.
    * `desired_user_sessions` - Desired number of user sessions for the WorkSpaces Pool.
* `description` - Description of the WorkSpaces Pool.
* `directory_id` - ID of the directory for the WorkSpaces Pool.
* `state` - Current state of the WorkSpaces Pool.
* `tags` - Map of tags assigned to the resource.
* `timeout_settings` - Information about the timeout settings for the WorkSpaces Pool.
    * `disconnect_timeout_in_seconds` - Time after disconnection when a user is logged out of their WorkSpace.
    * `idle_disconnect_timeout_in_seconds` - Time after inactivity when a user is disconnected from their WorkSpace.
    * `max_user_duration_in_seconds` - Maximum time that a user can be connected to their WorkSpace.
