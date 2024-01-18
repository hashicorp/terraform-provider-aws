---
subcategory: "Security Lake"
layout: "aws"
page_title: "AWS: aws_securitylake_custom_log_source"
description: |-
  Terraform resource for managing an AWS Security Lake Custom Log Source.
---

# Resource: aws_securitylake_custom_log_source

Terraform resource for managing an AWS Security Lake Custom Log Source.

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
}
```

## Argument Reference

The following arguments are required:

* `event_classes` - (Required) The Open Cybersecurity Schema Framework (OCSF) event classes which describes the type of data that the custom source will send to Security Lake.
* `configuration` - (Required) SThe configuration for the third-party custom source.
* `source_name` - (Required) Specify the name for a third-party custom source. This must be a Regionally unique value.
* `source_version` - (Optional) Specify the source version for the third-party custom source, to limit log collection to a specific version of custom data source.

Configurations support the following:

* `crawler_configuration` - (Required) The configuration for the Glue Crawler for the third-party custom source.
* `provider_identity` - (Optional) The identity of the log provider for the third-party custom source.

Crawler Configuration support the following:

* `role_arn` - (Required) The Amazon Resource Name (ARN) of the AWS Identity and Access Management (IAM) role to be used by the AWS Glue crawler.

Provider Identity support the following:

* `external_id` - (Required) The external ID used to estalish trust relationship with the AWS identity.
* `principal` - (Required) The AWS identity principal.

## Attribute Reference

This resource exports no additional attributes.

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
