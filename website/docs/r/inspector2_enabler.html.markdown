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

## Argument Reference

The following arguments are required:

* `resource_types` - (Required) Type of resources to scan. Valid values are `EC2` and `ECR`. If you only use one type, Terraform will ignore the status of the other type.

The following arguments are optional:

* `account_ids` - (Optional) Set of account IDs. The default is to enable scans on the account where the resource is used.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)
