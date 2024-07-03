---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_connector"
description: |-
  Terraform data source for managing an AWS Transfer Family Connector.
---

# Data Source: aws_transfer_connector

Terraform data source for managing an AWS Transfer Family Connector.

### Basic Usage

```terraform
data "aws_transfer_connector" "test" {
  id = "c-xxxxxxxxxxxxxx"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Unique identifier for connector

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `access_role` - ARN of the AWS Identity and Access Management role.
* `arn` - ARN of the Connector.
* `as2_config` - Structure containing the parameters for an AS2 connector object. Contains the following attributes:
    * `basic_auth_secret_id` -  Basic authentication for AS2 connector API. Returns a null value if not set.
    * `compression` - Specifies whether AS2 file is compressed. Will be ZLIB or DISABLED
    * `encryption_algorithm` - Algorithm used to encrypt file. Will be AES128_CBC or AES192_CBC or AES256_CBC or DES_EDE3_CBC or NONE.
    * `local_profile_id` - Unique identifier for AS2 local profile.
    * `mdn_response` - Used for outbound requests to tell if response is asynchronous or not. Will be either SYNC or NONE.
    * `mdn_signing_algorithm` - Signing algorithm for MDN response. Will be SHA256 or SHA384 or SHA512 or SHA1 or NONE or DEFAULT.
    * `message_subject` - Subject HTTP header attribute in outbound AS2 messages to the connector.
    * `partner_profile_id` - Unique identifier used by connector for partner profile.
    * `signing_algorithm` - Algorithm used for signing AS2 messages sent with the connector.
* `logging_role` -  ARN of the IAM role that allows a connector to turn on CLoudwatch logging for Amazon S3 events.
* `security_policy_name` - Name of security policy.
* `service_managed_egress_ip_addresses` - List of egress Ip addresses.
* `sftp_config` - Object containing the following attributes:
    * `trusted_host_keys` - List of the public portions of the host keys that are used to identify the servers the connector is connected to.
    * `user_secret_id` - Identifer for the secret in AWS Secrets Manager that contains the SFTP user's private key, and/or password.
* `tags` - Object containing the following attributes:
    * `key` - Name of the tag.
    * `value` - Values associated with the tags key.
* `url` - URL of the partner's AS2 or SFTP endpoint.
