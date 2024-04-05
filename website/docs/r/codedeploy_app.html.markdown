---
subcategory: "CodeDeploy"
layout: "aws"
page_title: "AWS: aws_codedeploy_app"
description: |-
  Provides a CodeDeploy application.
---

# Resource: aws_codedeploy_app

Provides a CodeDeploy application to be used as a basis for deployments

## Example Usage

### ECS Application

```terraform
resource "aws_codedeploy_app" "example" {
  compute_platform = "ECS"
  name             = "example"
}
```

### Lambda Application

```terraform
resource "aws_codedeploy_app" "example" {
  compute_platform = "Lambda"
  name             = "example"
}
```

### Server Application

```terraform
resource "aws_codedeploy_app" "example" {
  compute_platform = "Server"
  name             = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the application.
* `compute_platform` - (Optional) The compute platform can either be `ECS`, `Lambda`, or `Server`. Default is `Server`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the CodeDeploy application.
* `application_id` - The application ID.
* `id` - Amazon's assigned ID for the application.
* `name` - The application's name.
* `github_account_name` - The name for a connection to a GitHub account.
* `linked_to_github` - Whether the user has authenticated with GitHub for the specified application.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeDeploy Applications using the `name`. For example:

```terraform
import {
  to = aws_codedeploy_app.example
  id = "my-application"
}
```

Using `terraform import`, import CodeDeploy Applications using the `name`. For example:

```console
% terraform import aws_codedeploy_app.example my-application
```
