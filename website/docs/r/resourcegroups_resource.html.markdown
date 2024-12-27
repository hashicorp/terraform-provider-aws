---
subcategory: "Resource Groups"
layout: "aws"
page_title: "AWS: aws_resourcegroups_resource"
description: |-
  Terraform resource for managing an AWS Resource Groups Resource.
---

# Resource: aws_resourcegroups_resource

Terraform resource for managing an AWS Resource Groups Resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_host" "example" {
  instance_family   = "t3"
  availability_zone = "us-east-1a"
  host_recovery     = "off"
  auto_placement    = "on"
}

resource "aws_resourcegroups_group" "example" {
  name = "example"
}

resource "aws_resourcegroups_resource" "example" {
  group_arn    = aws_resourcegroups_group.example.arn
  resource_arn = aws_ec2_host.example.arn
}
```

## Argument Reference

The following arguments are required:

* `group_arn` - (Required) Name or ARN of the resource group to add resources to.
* `resource_arn` - (Required) ARN of the resource to be added to the group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A comma-delimited string combining `group_arn` and `resource_arn`.
* `resource_type` - The resource type of a resource, such as `AWS::EC2::Instance`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS Resource Groups Resource using `group_arn` and `resource_arn`, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_resourcegroups_resource.example
  id = "arn:aws:resource-groups:us-west-2:012345678901:group/example,arn:aws:lambda:us-west-2:012345678901:function:example"
}
```

Using `terraform import`, import an AWS Resource Groups Resource using `group_arn` and `resource_arn`, separated by a comma (`,`). For example:

```console
% terraform import aws_resourcegroups_resource.example arn:aws:resource-groups:us-west-2:012345678901:group/example,arn:aws:lambda:us-west-2:012345678901:function:example
```
