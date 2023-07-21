---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_agreement"
description: |-
  Provides a AWS Transfer AS2 Agreement Resource
---

# Resource: aws_transfer_agreement

Provides a AWS Transfer AS2 Agreement resource.

## Example Usage

### Basic

```terraform
resource "aws_transfer_agreement" "example" {
  access_role        = aws_iam_role.test.arn
  base_directory     = "/DOC-EXAMPLE-BUCKET/home/mydirectory"
  description        = "example"
  local_profile_id   = aws_transfer_profile.local.profile_id
  partner_profile_id = aws_transfer_profile.partner.profile_id
  server_id          = aws_transfer_server.test.id
}
```

## Argument Reference

This resource supports the following arguments:

* `access_role` - (Required) The IAM Role which provides read and write access to the parent directory of the file location mentioned in the StartFileTransfer request.
* `base_directory` - (Required) The landing directory for the files transferred by using the AS2 protocol.
* `description` - (Optional) The Optional description of the transdfer.
* `local_profile_id` - (Required) The unique identifier for the AS2 local profile.
* `partner_profile_id` - (Required) The unique identifier for the AS2 partner profile.
* `server_id` - (Required) The unique server identifier for the server instance. This is the specific server the agreement uses.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `agreement_id`  - The unique identifier for the AS2 agreement
* `staus`  - The staus of the agreement which is either ACTIVE or INACTIVE.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transfer AS2 Agreement using the `server_id/agreement_id`. For example:

```terraform
import {
  to = aws_transfer_agreement.example
  id = "s-4221a88afd5f4362a/a-4221a88afd5f4362a"
}
```

Using `terraform import`, import Transfer AS2 Agreement using the `server_id/agreement_id`. For example:

```console
% terraform import aws_transfer_agreement.example s-4221a88afd5f4362a/a-4221a88afd5f4362a
```
