---
subcategory: "Network Flow Monitor"
layout: "aws"
page_title: "AWS: aws_networkflowmonitor_scope"
description: |-
  Manages a Network Flow Monitor Scope.
---

# Resource: aws_networkflowmonitor_scope

Manages a Network Flow Monitor Scope.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}

resource "aws_networkflowmonitor_scope" "example" {
  targets {
    region = "us-east-1"
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }

  tags = {
    Name = "example"
  }
}
```


## Argument Reference

This resource supports the following arguments:

* `targets` - (Required) The targets to define the scope to be monitored. A target is an array of target resources, which are currently Region-account pairs.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### targets

The `targets` block supports the following:

* `region` - (Required) The AWS Region for the scope.
* `target_identifier` - (Required) A target identifier is a pair of identifying information for a scope.

### target_identifier

The `target_identifier` block supports the following:

* `target_id` - (Required) The identifier for a target, which is currently always an account ID.
* `target_type` - (Required) The type of a target. A target type is currently always `ACCOUNT`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the scope.
* `id` - The Amazon Resource Name (ARN) of the scope.
* `scope_id` - The identifier for the scope that includes the resources you want to get data results for.
* `status` - The status for a scope. The status can be one of the following: `SUCCEEDED`, `IN_PROGRESS`, `FAILED`, `DEACTIVATING`, or `DEACTIVATED`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Flow Monitor Scope using the scope ARN. For example:

```terraform
import {
  to = aws_networkflowmonitor_scope.example
  id = "arn:aws:networkflowmonitor:us-east-1:123456789012:scope/example-scope-id"
}
```

Using `terraform import`, import Network Flow Monitor Scope using the scope ARN. For example:

```console
% terraform import aws_networkflowmonitor_scope.example arn:aws:networkflowmonitor:us-east-1:123456789012:scope/example-scope-id
```