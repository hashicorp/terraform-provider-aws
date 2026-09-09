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

### SFTP Connector with VPC Lattice

```terraform
resource "aws_transfer_connector" "example" {
  access_role = aws_iam_role.test.arn
  sftp_config {
    trusted_host_keys = ["ssh-rsa AAAAB3NYourKeysHere"]
    user_secret_id    = aws_secretsmanager_secret.example.id
  }
  egress_config {
    vpc_lattice {
      resource_configuration_arn = "arn:aws:vpc-lattice:us-east-1:123456789012:resourceconfiguration/rcfg-12345678901234567"
      port_number                = 22
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `access_role` - (Required) IAM Role which provides read and write access to the parent directory of the file location mentioned in the StartFileTransfer request.
* `as2_config` - (Optional) Either SFTP or AS2 is configured. Parameters to configure for the connector object. See [`as2_config` Block](#as2_config-block) below.
* `egress_config` - (Optional) Egress configuration for the connector. When set, enables routing through customer VPCs using VPC Lattice for private connectivity. See [`egress_config` Block](#egress_config-block) below.
* `logging_role` - (Optional) IAM Role which is required for allowing the connector to turn on CloudWatch logging for Amazon S3 events.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `security_policy_name` - (Optional) Name of the security policy for the connector.
* `sftp_config` - (Optional) Either SFTP or AS2 is configured. Parameters to configure for the connector object. See [`sftp_config` Block](#sftp_config-block) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `url` - (Optional) URL of the partners AS2 endpoint or SFTP endpoint. Required for AS2 connectors and service-managed SFTP connectors. Must be null when using VPC Lattice egress configuration.

### `as2_config` Block

* `compression` - (Required) Whether AS2 file is compressed. The valid values are ZLIB and DISABLED.
* `encryption_algorithm` - (Required) Algorithm that is used to encrypt the file. The valid values are AES128_CBC | AES192_CBC | AES256_CBC | NONE.
* `local_profile_id` - (Required) Unique identifier for the AS2 local profile.
* `mdn_response` - (Required) Determines, for outbound requests, if a partner response for transfers is synchronous or asynchronous. The valid values are SYNC and NONE.
* `mdn_signing_algorithm` - (Optional) Signing algorithm for the MDN response. The valid values are SHA256 | SHA384 | SHA512 | SHA1 | NONE | DEFAULT.
* `message_subject` - (Optional) Subject HTTP header attribute used in AS2 messages that are being sent with the connector.
* `partner_profile_id` - (Required) Unique identifier for the AS2 partner profile.
* `signing_algorithm` - (Required) Algorithm that is used to sign AS2 messages sent with the connector. The valid values are SHA256 | SHA384 | SHA512 | SHA1 | NONE .

### `sftp_config` Block

* `trusted_host_keys` - (Required) List of public portion of the host key, or keys, that are used to authenticate the user to the external server to which you are connecting.(https://docs.aws.amazon.com/transfer/latest/userguide/API_SftpConnectorConfig.html)
* `user_secret_id` - (Required) Identifier for the secret (in AWS Secrets Manager) that contains the SFTP user's private key, password, or both. The identifier can be either the ARN or the name of the secret.

### `egress_config` Block

* `vpc_lattice` - (Optional) VPC Lattice configuration for routing connector traffic through customer VPCs. See [`vpc_lattice` Block](#vpc_lattice-block) below.

### `vpc_lattice` Block

* `port_number` - (Optional) Port number for connecting to the SFTP server through VPC Lattice. Defaults to 22 if not specified. Must match the port on which the target SFTP server is listening. Valid values are between 1 and 65535.
* `resource_configuration_arn` - (Required) ARN of the VPC Lattice Resource Configuration that defines the target SFTP server location. Must point to a valid Resource Configuration in a VPC with appropriate network connectivity to the SFTP server.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connector.
* `connector_id` - Unique identifier for the AS2 profile or SFTP Profile.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

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
