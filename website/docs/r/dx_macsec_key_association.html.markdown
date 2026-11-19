---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_macsec_key_association"
description: Manages a MAC Security (MACSec) secret key association for use with Direct Connect.
---

# Resource: aws_dx_macsec_key_association

Manages a MAC Security (MACSec) secret key association for use with Direct Connect. See [MACsec prerequisites](https://docs.aws.amazon.com/directconnect/latest/UserGuide/direct-connect-mac-sec-getting-started.html#mac-sec-prerequisites) for information about MAC Security (MACsec) prerequisites.

Creating this resource will also create a resource of type [`aws_secretsmanager_secret`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret) which is managed by Direct Connect. While you can import this resource into your Terraform state, because this secret is managed by Direct Connect, you will not be able to make any modifications to it. See [How AWS Direct Connect uses AWS Secrets Manager](https://docs.aws.amazon.com/secretsmanager/latest/userguide/integrating_how-services-use-secrets_directconnect.html) for details.

~> **Note:** The `ckn` and `cak` arguments are stored in the raw state as plain-text. Use `ckn_wo` and `cak_wo` instead to avoid storing these values in the plan or state. When `ckn_wo` is configured, the API-returned CKN is not stored in `ckn`.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

-> **Note:** Write-only arguments `cak_wo` and `ckn_wo` require Terraform v1.11.0 or later. [Learn more about write-only arguments](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments).

~> **Note:** The `secret_arn` argument can only be used to reference a previously created MACSec key. You cannot associate a Secrets Manager secret created outside of the `aws_dx_macsec_key_association` resource.

## Example Usage

### Basic Usage

```terraform
data "aws_dx_connection" "example" {
  name = "tf-dx-connection"
}

resource "aws_dx_macsec_key_association" "example" {
  connection_id = data.aws_dx_connection.example.id
  ckn           = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
  cak           = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
}
```

### Create MACSec key with write-only CKN and CAK

```terraform
data "aws_dx_connection" "example" {
  name = "tf-dx-connection"
}

variable "cak" {
  type      = string
  sensitive = true
  ephemeral = true
}

variable "ckn" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "aws_dx_macsec_key_association" "example" {
  connection_id  = data.aws_dx_connection.example.id
  cak_wo         = var.cak
  cak_wo_version = 1
  ckn_wo         = var.ckn
  ckn_wo_version = 1
}
```

Changing a write-only value alone does not trigger a change. Increment its corresponding `cak_wo_version` or `ckn_wo_version` to replace the association with the new key material. The version counters are independent; increment both when changing both values.

### Create MACSec key with existing Secrets Manager secret

```terraform
data "aws_dx_connection" "example" {
  name = "tf-dx-connection"
}

data "aws_secretsmanager_secret" "example" {
  name = "directconnect!prod/us-east-1/directconnect/0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
}

resource "aws_dx_macsec_key_association" "example" {
  connection_id = data.aws_dx_connection.example.id
  secret_arn    = data.aws_secretsmanager_secret.example.arn
}
```

## Argument Reference

The following arguments are required:

* `connection_id` - (Required) ID of the dedicated Direct Connect connection. The connection must be a dedicated connection in the `AVAILABLE` state.

The following arguments are optional:

* `cak` - (Optional) MAC Security (MACsec) CAK to associate with the dedicated connection. Valid values are 64 hexadecimal characters (0-9, A-F). Required with `ckn` or `ckn_wo` unless `cak_wo` is configured. Conflicts with `cak_wo`. Changing this argument replaces the association.
* `cak_wo` - (Optional) Write-only MAC Security (MACsec) CAK to associate with the dedicated connection. Valid values are 64 hexadecimal characters (0-9, A-F). Requires `cak_wo_version` and either `ckn` or `ckn_wo`. Conflicts with `cak` and `secret_arn`. This value is not stored in the plan or state.
* `cak_wo_version` - (Optional) Positive integer used to trigger replacement when `cak_wo` changes. Required with `cak_wo`. Increment this value to replace the association using the current write-only values.
* `ckn` - (Optional) MAC Security (MACsec) CKN to associate with the dedicated connection. Valid values are 64 hexadecimal characters (0-9, A-F). Required with `cak` or `cak_wo` unless `ckn_wo` is configured. Conflicts with `ckn_wo`. Changing this argument replaces the association.
* `ckn_wo` - (Optional) Write-only MAC Security (MACsec) CKN to associate with the dedicated connection. Valid values are 64 hexadecimal characters (0-9, A-F). Requires `ckn_wo_version` and either `cak` or `cak_wo`. Conflicts with `ckn` and `secret_arn`. This value is not stored in the plan or state.
* `ckn_wo_version` - (Optional) Positive integer used to trigger replacement when `ckn_wo` changes. Required with `ckn_wo`. Increment this value to replace the association using the current write-only values.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `secret_arn` - (Optional) ARN of the MAC Security (MACsec) secret key to associate with the dedicated connection.

~> **Note:** Supply either a CAK/CKN pair (using regular or write-only arguments) or `secret_arn`. Do not combine key arguments with `secret_arn`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the MAC Security (MACSec) secret key resource.
* `start_on` - The date in UTC format that the MAC Security (MACsec) secret key takes effect.
* `state` -  The state of the MAC Security (MACsec) secret key. The possible values are: associating, associated, disassociating, disassociated. See [MacSecKey](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_MacSecKey.html#DX-Type-MacSecKey-state) for descriptions of each state.
