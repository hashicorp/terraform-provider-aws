---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_macsec_key_association"
description: |-
  Provides a MAC Security (MACSec) secret key resource for use with Direct Connect.
---

# Resource: aws_dx_macsec_key_association

Provides a MAC Security (MACSec) secret key resource for use with Direct Connect. See [MACsec prerequisites](https://docs.aws.amazon.com/directconnect/latest/UserGuide/direct-connect-mac-sec-getting-started.html#mac-sec-prerequisites) for information about MAC Security (MACsec) prerequisites.

Creating this resource will also create a resource of type [`aws_secretsmanager_secret`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret) which is managed by Direct Connect. While you can import this resource into your Terraform state, because this secret is managed by Direct Connect, you will not be able to make any modifications to it. See [How AWS Direct Connect uses AWS Secrets Manager](https://docs.aws.amazon.com/secretsmanager/latest/userguide/integrating_how-services-use-secrets_directconnect.html) for details.

~> **Note:** All arguments including `ckn` and `cak` will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

~> **Note:** The `secret_arn` argument can only be used to reference a previously created MACSec key. You cannot associate a Secrets Manager secret created outside of the `aws_dx_macsec_key_association` resource.

## Example Usage

### Create MACSec key with CKN and CAK

```terraform
data "aws_dx_connection" "example" {
  name = "tf-dx-connection"
}

resource "aws_dx_macsec_key_association" "test" {
  connection_id = data.aws_dx_connection.example.id
  ckn           = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
  cak           = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
}
```

### Create MACSec key with existing Secrets Manager secret

```terraform
data "aws_dx_connection" "example" {
  name = "tf-dx-connection"
}

data "aws_secretsmanager_secret" "example" {
  name = "directconnect!prod/us-east-1/directconnect/0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
}

resource "aws_dx_macsec_key_association" "test" {
  connection_id = data.aws_dx_connection.example.id
  secret_arn    = data.aws_secretsmanager_secret.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `cak` - (Optional) The MAC Security (MACsec) CAK to associate with the dedicated connection. The valid values are 64 hexadecimal characters (0-9, A-E). Required if using `ckn`.
* `ckn` - (Optional) The MAC Security (MACsec) CKN to associate with the dedicated connection. The valid values are 64 hexadecimal characters (0-9, A-E). Required if using `cak`.
* `connection_id` - (Required) The ID of the dedicated Direct Connect connection. The connection must be a dedicated connection in the `AVAILABLE` state.
* `secret_arn` - (Optional) The Amazon Resource Name (ARN) of the MAC Security (MACsec) secret key to associate with the dedicated connection.

~> **Note:** `ckn` and `cak` are mutually exclusive with `secret_arn` - these arguments cannot be used together. If you use `ckn` and `cak`, you should not use `secret_arn`. If you use the `secret_arn` argument to reference an existing MAC Security (MACSec) secret key, you should not use `ckn` or `cak`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the MAC Security (MACSec) secret key resource.
* `start_on` - The date in UTC format that the MAC Security (MACsec) secret key takes effect.
* `state` -  The state of the MAC Security (MACsec) secret key. The possible values are: associating, associated, disassociating, disassociated. See [MacSecKey](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_MacSecKey.html#DX-Type-MacSecKey-state) for descriptions of each state.
