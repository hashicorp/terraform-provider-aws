---
subcategory: "Verified Access"
layout: "aws"
page_title: "AWS: aws_verifiedaccess_group"
description: |-
  Terraform resource for managing a Verified Access Group.
---

# Resource: aws_verifiedaccess_group

Terraform resource for managing a Verified Access Group.

## Example Usage

```terraform
resource "aws_verifiedaccess_group" "example" {
  description                 = "example"
  verified_access_instance_id = aws_verifiedaccess_instance.example.id

  depends_on = [
    aws_verifiedaccess_trust_provider_attachment.example
  ]
}

resource "aws_verifiedaccess_instance" "example" {
  description = "example"
}

resource "aws_verifiedaccess_trust_provider" "example" {
  policy_reference_name    = "example"
  trust_provider_type      = "user"
  user_trust_provider_type = "iam-identity-center"
}

resource "aws_verifiedaccess_trust_provider_attachment" "example" {
  verified_access_instance_id       = aws_verifiedaccess_instance.example.id
  verified_access_trust_provider_id = aws_verifiedaccess_trust_provider.example.id
}
```

## Argument Reference

The following arguments are required:

* `verified_access_instance_id` - (Required) The ID of the AWS Verified Access instance. Note: The Instance must have a Trust Provider attached before you can create a Group.

The following arguments are optional:

* `description` - (Optional) A description for the AWS Verified Access group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Verified Access group.
* `id` - The ID of the Verified Access group.
* `owner` - The AWS account number that owns the group.
* `verified_access_instance_id` - The ID of the Verified Access instance.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

Verified Access Groups can be imported using the `id`, e.g.,

```
$ terraform import aws_verifiedaccess_group.example vagr-8012925589
```
