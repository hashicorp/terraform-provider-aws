---
subcategory: "Verified Access"
layout: "aws"
page_title: "AWS: aws_verifiedaccess_instance"
description: |-
  Terraform resource for managing a Verified Access Instance.
---

# Resource: aws_verifiedaccess_instance

Terraform resource for managing a Verified Access Instance.

## Example Usage

```terraform
resource "aws_verifiedaccess_instance" "example" {
  description = "example"
}
```

## Argument Reference

The following arguments are optional:

* `description` - (Optional) A description for the AWS Verified Access instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the AWS Verified Access instance.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

Verified Access Instances can be imported using the `id`, e.g.,

```
$ terraform import aws_verifiedaccess_instance.example vai-8012925589
```
