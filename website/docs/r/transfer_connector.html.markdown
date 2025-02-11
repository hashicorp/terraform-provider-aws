---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_connector"
description: |-
  Provides a AWS Transfer AS2 Connector Resource
---

# Resource: aws_transfer_connector

Provides a AWS Transfer AS2 Connector resource.

## Example Usage

### Basic

```terraform
resource "aws_transfer_connector" "example" {
  access_role = aws_iam_role.test.arn
  as2_config {
    compression           = "DISABLED"
    encryption_algorithm  = "AWS128_CBC"
    message_subject       = "For Connector"
    local_profile_id      = aws_transfer_profile.local.profile_id
    mdn_response          = "NONE"
    mdn_signing_algorithm = "NONE"
    partner_profile_id    = aws_transfer_profile.partner.profile_id
    signing_algorithm     = "NONE"
  }
  url = "http://www.test.com"
}
```

### SFTP Connector

```terraform
resource "aws_transfer_connector" "example" {
  access_role = aws_iam_role.test.arn
  sftp_config {
    trusted_host_keys = ["ssh-rsa AAAAB3NYourKeysHere"]
    user_secret_id    = aws_secretsmanager_secret.example.id
  }
  url = "sftp://test.com"
}
```

## Argument Reference

This resource supports the following arguments:

* `access_role` - (Required) The IAM Role which provides read and write access to the parent directory of the file location mentioned in the StartFileTransfer request.
* `as2_config` - (Optional) Either SFTP or AS2 is configured.The parameters to configure for the connector object. Fields documented below.
* `logging_role` - (Optional) The IAM Role which is required for allowing the connector to turn on CloudWatch logging for Amazon S3 events.
* `security_policy_name` - (Optional) Name of the security policy for the connector.
* `sftp_config` - (Optional) Either SFTP or AS2 is configured.The parameters to configure for the connector object. Fields documented below.
* `url` - (Required) The URL of the partners AS2 endpoint or SFTP endpoint.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### As2Config Details

* `compression` - (Required) Specifies weather AS2 file is compressed. The valud values are ZLIB and  DISABLED.
* `encryption_algorithm` - (Required) The algorithm that is used to encrypt the file. The valid values are AES128_CBC | AES192_CBC | AES256_CBC | NONE.
* `local_profile_id` - (Required) The unique identifier for the AS2 local profile.
* `mdn_response` - (Required) Used for outbound requests to determine if a partner response for transfers is synchronous or asynchronous. The valid values are SYNC and NONE.
* `mdn_signing_algorithm` - (Optional) The signing algorithm for the Mdn response. The valid values are SHA256 | SHA384 | SHA512 | SHA1 | NONE | DEFAULT.
* `message_subject` - (Optional) Used as the subject HTTP header attribute in AS2 messages that are being sent with the connector.
* `partner_profile_id` - (Required) The unique identifier for the AS2 partner profile.
* `signing_algorithm` - (Required) The algorithm that is used to sign AS2 messages sent with the connector. The valid values are SHA256 | SHA384 | SHA512 | SHA1 | NONE .

### SftpConfig Details

* `trusted_host_keys` - (Required) A list of public portion of the host key, or keys, that are used to authenticate the user to the external server to which you are connecting.(https://docs.aws.amazon.com/transfer/latest/userguide/API_SftpConnectorConfig.html)
* `user_secret_id` - (Required) The identifier for the secret (in AWS Secrets Manager) that contains the SFTP user's private key, password, or both. The identifier can be either the Amazon Resource Name (ARN) or the name of the secret.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the connector.
* `connector_id`  - The unique identifier for the AS2 profile or SFTP Profile.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transfer AS2 Connector using the `connector_id`. For example:

```terraform
import {
  to = aws_transfer_connector.example
  id = "c-4221a88afd5f4362a"
}
```

Using `terraform import`, import Transfer AS2 Connector using the `connector_id`. For example:

```console
% terraform import aws_transfer_connector.example c-4221a88afd5f4362a
```
