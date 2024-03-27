---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_user_login_profile"
description: |-
  Manages an IAM User Login Profile
---

# Resource: aws_iam_user_login_profile

Manages an IAM User Login Profile with limited support for password creation during Terraform resource creation. Uses PGP to encrypt the password for safe transport to the user. PGP keys can be obtained from Keybase.

-> To reset an IAM User login password via Terraform, you can use the [`terraform taint` command](https://www.terraform.io/docs/commands/taint.html) or change any of the arguments.

## Example Usage

```terraform
resource "aws_iam_user" "example" {
  name          = "example"
  path          = "/"
  force_destroy = true
}

resource "aws_iam_user_login_profile" "example" {
  user    = aws_iam_user.example.name
  pgp_key = "keybase:some_person_that_exists"
}

output "password" {
  value = aws_iam_user_login_profile.example.encrypted_password
}
```

## Argument Reference

This resource supports the following arguments:

* `user` - (Required) The IAM user's name.
* `pgp_key` - (Optional) Either a base-64 encoded PGP public key, or a keybase username in the form `keybase:username`. Only applies on resource creation. Drift detection is not possible with this argument.
* `password_length` - (Optional) The length of the generated password on resource creation. Only applies on resource creation. Drift detection is not possible with this argument. Default value is `20`.
* `password_reset_required` - (Optional) Whether the user should be forced to reset the generated password. This is applied all the time! If you want to avoid the situation where you set up a new user, they change their password as required, then the next `terraform apply` causes that user to be destroyed and then created with a new initial password, then, you need to use:

  ```
  lifecycle {
    ignore_changes = [
      password_reset_required,
    ]
  }
  ```

  Currently waiting on issue [#23567](https://github.com/hashicorp/terraform-provider-aws/issues/23567) to be prioritized which would hope to restore the default behaviour of applying this only on resource creation.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `password` - The plain text password, only available when `pgp_key` is not provided.
* `key_fingerprint` - The fingerprint of the PGP key used to encrypt the password. Only available if password was handled on Terraform resource creation, not import.
* `encrypted_password` - The encrypted password, base64 encoded. Only available if password was handled on Terraform resource creation, not import.

~> **NOTE:** The encrypted password may be decrypted using the command line,
   for example: `terraform output password | base64 --decode | keybase pgp decrypt`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM User Login Profiles without password information via the IAM User name. For example:

```terraform
import {
  to = aws_iam_user_login_profile.example
  id = "myusername"
}
```

Using `terraform import`, import IAM User Login Profiles without password information via the IAM User name. For example:

```console
% terraform import aws_iam_user_login_profile.example myusername
```

Since Terraform has no method to read the PGP or password information during import, use the [Terraform resource `lifecycle` configuration block `ignore_changes` argument](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to ignore them (unless you want to recreate a password). For example:

```terraform
resource "aws_iam_user_login_profile" "example" {
  # ... other configuration ...

  lifecycle {
    ignore_changes = [
      password_length,
      password_reset_required,
      pgp_key,
    ]
  }
}
```
