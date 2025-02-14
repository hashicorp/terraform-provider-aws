---
subcategory: "Security Lake"
layout: "aws"
page_title: "AWS: aws_securitylake_subscriber"
description: |-
  Terraform resource for managing an AWS Security Lake Subscriber.
---

# Resource: aws_securitylake_subscriber

Terraform resource for managing an AWS Security Lake Subscriber.

~> **NOTE:** The underlying `aws_securitylake_data_lake` must be configured before creating the `aws_securitylake_subscriber`. Use a `depends_on` statement.

## Example Usage

```terraform
resource "aws_securitylake_subscriber" "example" {
  subscriber_name = "example-name"
  access_type     = "S3"

  source {
    aws_log_source_resource {
      source_name    = "ROUTE53"
      source_version = "1.0"
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = "1234567890"
  }

  depends_on = [aws_securitylake_data_lake.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `access_type` - (Optional) The Amazon S3 or Lake Formation access type.
* `source` - (Required) The supported AWS services from which logs and events are collected. Security Lake supports log and event collection for natively supported AWS services. See [`source` Blocks](#source-blocks) below.
* `subscriber_identity` - (Required) The AWS identity used to access your data. See [`subscriber_identity` Block](#subscriber_identity-block) below.
* `subscriber_description` - (Optional) The description for your subscriber account in Security Lake.
* `subscriber_name` - (Optional) The name of your Security Lake subscriber account.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `source` Blocks

`source` blocks support the following arguments:

* `aws_log_source_resource` - (Optional) Amazon Security Lake supports log and event collection for natively supported AWS services. See [`aws_log_source_resource` Block](#aws_log_source_resource-block) below.
* `custom_log_source_resource` - (Optional) Amazon Security Lake supports custom source types. See [`custom_log_source_resource` Block](#custom_log_source_resource-block) below.

### `subscriber_identity` Block

The `subscriber_identity` block supports the following arguments:

* `external_id` - (Required) The AWS Regions where Security Lake is automatically enabled.
* `principal` - (Required) Provides encryption details of Amazon Security Lake object.

### `aws_log_source_resource` Block

The `aws_log_source_resource` block supports the following arguments:

* `source_name` - (Required) Provides data expiration details of Amazon Security Lake object.
* `source_version` - (Optional) Provides data storage transition details of Amazon Security Lake object.

### `custom_log_source_resource` Block

The `custom_log_source_resource` block supports the following arguments. See [`custom_log_source_resource` Attribute Reference](#custom_log_source_resource-attribute-reference) below for additional read-only attributes.

* `source_name` - (Required) The name for a third-party custom source. This must be a Regionally unique value.
* `source_version` - (Optional) The version for a third-party custom source. This must be a Regionally unique value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Data Lake.
* `id` - The Subscriber ID of the subscriber.
* `s3_bucket_arn` - The ARN for the Amazon Security Lake Amazon S3 bucket.
* `resource_share_arn` - The Amazon Resource Name (ARN) which uniquely defines the AWS RAM resource share. Before accepting the RAM resource share invitation, you can view details related to the RAM resource share.
* `role_arn` - The Amazon Resource Name (ARN) specifying the role of the subscriber.
* `subscriber_endpoint` - The subscriber endpoint to which exception messages are posted.
* `subscriber_status` - The subscriber status of the Amazon Security Lake subscriber account.
* `resource_share_name` - The name of the resource share.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `custom_log_source_resource` Attribute Reference

The `custom_log_source_resource` block exports the following attributes in addition to the arguments above:

* `attributes` - The attributes of the third-party custom source. See [`attributes` Block](#attributes-block) below.
* `provider` - The details of the log provider for the third-party custom source. See [`provider` Block](#provider-block) below.

### `attributes` Block

The `attributes` block exports the following attributes:

* `crawler_arn` - The ARN of the AWS Glue crawler.
* `database_arn` - The ARN of the AWS Glue database where results are written.
* `table_arn` - The ARN of the AWS Glue table.

### `provider` Block

The `provider` block exports the following attributes:

* `location` - The location of the partition in the Amazon S3 bucket for Security Lake.
* `role_arn` - The ARN of the IAM role to be used by the entity putting logs into your custom source partition.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Lake subscriber using the subscriber ID. For example:

```terraform
import {
  to = aws_securitylake_subscriber.example
  id = "9f3bfe79-d543-474d-a93c-f3846805d208"
}
```

Using `terraform import`, import Security Lake subscriber using the subscriber ID. For example:

```console
% terraform import aws_securitylake_subscriber.example 9f3bfe79-d543-474d-a93c-f3846805d208
```
