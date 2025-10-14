---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_shared_directory_accepter"
description: |-
    Accepts a shared directory in a consumer account.
---

# Resource: aws_directory_service_shared_directory_accepter

Accepts a shared directory in a consumer account.

~> **NOTE:** Destroying this resource removes the shared directory from the consumer account only.

## Example Usage

```terraform
resource "aws_directory_service_shared_directory" "example" {
  directory_id = aws_directory_service_directory.example.id
  notes        = "example"

  target {
    id = data.aws_caller_identity.receiver.account_id
  }
}

resource "aws_directory_service_shared_directory_accepter" "example" {
  provider = "awsalternate"

  shared_directory_id = aws_directory_service_shared_directory.example.shared_directory_id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `shared_directory_id` - (Required) Identifier of the directory that is stored in the directory consumer account that corresponds to the shared directory in the owner account.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the shared directory.
* `method` - Method used when sharing a directory (i.e., `ORGANIZATIONS` or `HANDSHAKE`).
* `notes` - Message sent by the directory owner to the directory consumer to help the directory consumer administrator determine whether to approve or reject the share invitation.
* `owner_account_id` - Account identifier of the directory owner.
* `owner_directory_id` - Identifier of the Managed Microsoft AD directory from the perspective of the directory owner.

## Timeouts

`aws_directory_service_shared_directory_accepter` provides the following [Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) configuration options:

- `create` - (Default `60 minutes`) Used for directory creation
- `delete` - (Default `60 minutes`) Used for directory deletion

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Directory Service Shared Directories using the shared directory ID. For example:

```terraform
import {
  to = aws_directory_service_shared_directory_accepter.example
  id = "d-9267633ece"
}
```

Using `terraform import`, import Directory Service Shared Directories using the shared directory ID. For example:

```console
% terraform import aws_directory_service_shared_directory_accepter.example d-9267633ece
```
