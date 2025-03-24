---
subcategory: "Security Lake"
layout: "aws"
page_title: "AWS: aws_securitylake_custom_log_source"
description: |-
  Terraform resource for managing an AWS Security Lake Custom Log Source.
---

# Resource: aws_securitylake_custom_log_source

Terraform resource for managing an AWS Security Lake Custom Log Source.

~> **NOTE:** The underlying `aws_securitylake_data_lake` must be configured before creating the `aws_securitylake_custom_log_source`. Use a `depends_on` statement.

## Example Usage

### Basic Usage

```terraform
resource "aws_securitylake_custom_log_source" "example" {
  source_name    = "example-name"
  source_version = "1.0"
  event_classes  = ["FILE_ACTIVITY"]

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.custom_log.arn
    }

    provider_identity {
      external_id = "example-id"
      principal   = "123456789012"
    }
  }

  depends_on = [aws_securitylake_data_lake.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `configuration` - (Required) The configuration for the third-party custom source.
    * `crawler_configuration` - (Required) The configuration for the Glue Crawler for the third-party custom source.
        * `role_arn` - (Required) The Amazon Resource Name (ARN) of the AWS Identity and Access Management (IAM) role to be used by the AWS Glue crawler.
    * `provider_identity` - (Required) The identity of the log provider for the third-party custom source.
        * `external_id` - (Required) The external ID used to estalish trust relationship with the AWS identity.
        * `principal` - (Required) The AWS identity principal.
* `event_classes` - (Optional) The Open Cybersecurity Schema Framework (OCSF) event classes which describes the type of data that the custom source will send to Security Lake.
* `source_name` - (Required) Specify the name for a third-party custom source.
  This must be a Regionally unique value.
  Has a maximum length of 20.
* `source_version` - (Optional) Specify the source version for the third-party custom source, to limit log collection to a specific version of custom data source.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `attributes` - The attributes of a third-party custom source.
    * `crawler_arn` - The ARN of the AWS Glue crawler.
    * `database_arn` - The ARN of the AWS Glue database where results are written.
    * `table_arn` - The ARN of the AWS Glue table.
* `provider_details` - The details of the log provider for a third-party custom source.
    * `location` - The location of the partition in the Amazon S3 bucket for Security Lake.
    * `role_arn` - The ARN of the IAM role to be used by the entity putting logs into your custom source partition.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS log sources using the source name. For example:

```terraform
import {
  to = aws_securitylake_custom_log_source.example
  id = "example-name"
}
```

Using `terraform import`, import Custom log sources using the source name. For example:

```console
% terraform import aws_securitylake_custom_log_source.example example-name
```
