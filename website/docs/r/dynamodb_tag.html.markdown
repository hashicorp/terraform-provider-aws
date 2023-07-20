---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_tag"
description: |-
  Manages an individual DynamoDB resource tag
---

# Resource: aws_dynamodb_tag

Manages an individual DynamoDB resource tag. This resource should only be used in cases where DynamoDB resources are created outside Terraform (e.g., Table replicas in other regions).

~> **NOTE:** This tagging resource should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_dynamodb_table` and `aws_dynamodb_tag` to manage tags of the same DynamoDB Table in the same region will cause a perpetual difference where the `aws_dynamodb_cluster` resource will try to remove the tag being added by the `aws_dynamodb_tag` resource.

~> **NOTE:** This tagging resource does not use the [provider `ignore_tags` configuration](/docs/providers/aws/index.html#ignore_tags).

## Example Usage

```terraform
provider "aws" {
  region = "us-west-2"
}

provider "aws" {
  alias  = "replica"
  region = "us-east-1"
}

data "aws_region" "replica" {
  provider = aws.replica
}

data "aws_region" "current" {}

resource "aws_dynamodb_table" "example" {
  # ... other configuration ...

  replica {
    region_name = data.aws_region.replica.name
  }
}

resource "aws_dynamodb_tag" "test" {
  provider = aws.replica

  resource_arn = replace(aws_dynamodb_table.test.arn, data.aws_region.current.name, data.aws_region.replica.name)
  key          = "testkey"
  value        = "testvalue"
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required) Amazon Resource Name (ARN) of the DynamoDB resource to tag.
* `key` - (Required) Tag name.
* `value` - (Required) Tag value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - DynamoDB resource identifier and key, separated by a comma (`,`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_dynamodb_tag` using the DynamoDB resource identifier and key, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_dynamodb_tag.example
  id = "arn:aws:dynamodb:us-east-1:123456789012:table/example,Name"
}
```

Using `terraform import`, import `aws_dynamodb_tag` using the DynamoDB resource identifier and key, separated by a comma (`,`). For example:

```console
% terraform import aws_dynamodb_tag.example arn:aws:dynamodb:us-east-1:123456789012:table/example,Name
```
