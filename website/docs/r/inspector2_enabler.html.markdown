---
subcategory: "Inspector V2"
layout: "aws"
page_title: "AWS: aws_inspector2_enabler"
description: |-
  Terraform resource for enabling AWS Inspector V2 resource scans.
---

# Resource: aws_inspector2_enabler

Terraform resource for enabling AWS Inspector V2 resource scans.

~> **NOTE:** Due to testing limitations, we provide this resource as best effort. If you use it or have the ability to test it, and notice problems, please consider reaching out to us on [GitHub](https://github.com/hashicorp/terraform-provider-aws/issues/new/choose).

## Example Usage

### Basic Usage

```terraform
resource "aws_inspector2_enabler" "example" {
  account_ids    = ["012345678901"]
  resource_types = ["EC2"]
}
```

### For the Calling Account

```terraform
data "aws_caller_identity" "current" {}

resource "aws_inspector2_enabler" "test" {
  account_ids    = [data.aws_caller_identity.current.account_id]
  resource_types = ["ECR", "EC2"]
}
```

## Argument Reference

The following arguments are required:

* `account_ids` - (Required) Set of account IDs.
* `resource_types` - (Required) Type of resources to scan. Valid values are `EC2`, `ECR`, and `LAMBDA`. If you only use one type, Terraform will ignore the status of the other type.

## Attributes Reference

No additional attributes are exported.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)
