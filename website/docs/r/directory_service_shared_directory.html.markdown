---
subcategory: "DS (Directory Service)"
layout: "aws"
page_title: "AWS: aws_directory_service_shared_directory"
description: |-
    Manages a directory in your account (directory owner) shared with another account (directory consumer).
---

# Resource: aws_directory_service_shared_directory

Manages a directory in your account (directory owner) shared with another account (directory consumer).

## Example Usage

```terraform
resource "aws_directory_service_directory" "example" {
  name     = "tf-example"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.example.id
    subnet_ids = aws_subnet.example[*].id
  }
}

resource "aws_directory_service_shared_directory" "example" {
  directory_id = aws_directory_service_directory.example.id
  notes        = "You wanna have a catch?"

  target {
    id = data.aws_caller_identity.receiver.account_id
  }
}
```

## Argument Reference

The following arguments are required:

* `directory_id` - (Required) Identifier of the Managed Microsoft AD directory that you want to share with other accounts.
* `target` - (Required) Identifier for the directory consumer account with whom the directory is to be shared. See below.

The following arguments are optional:

* `method` - (Optional) Method used when sharing a directory. Valid values are `ORGANIZATIONS` and `HANDSHAKE`. Default is `HANDSHAKE`.
* `notes` - (Optional, Sensitive) Message sent by the directory owner to the directory consumer to help the directory consumer administrator determine whether to approve or reject the share invitation.

### `target`

* `id` - (Required) Identifier of the directory consumer account.
* `type` - (Optional) Type of identifier to be used in the `id` field. Valid value is `ACCOUNT`. Default is `ACCOUNT`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the shared directory.
* `shared_directory_id` - Identifier of the directory that is stored in the directory consumer account that corresponds to the shared directory in the owner account.

## Timeouts

`aws_directory_service_shared_directory` provides the following [Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) configuration options:

- `delete` - (Default `60 minutes`) Used for shared directory deletion

## Import

Directory Service Shared Directories can be imported using the owner directory ID/shared directory ID, e.g.,

```
$ terraform import aws_directory_service_shared_directory.example d-1234567890/d-9267633ece
```
