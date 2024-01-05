---
subcategory: "Security Lake"
layout: "aws"
page_title: "AWS: aws_securitylake_aws_log_source"
description: |-
  Terraform resource for managing an Amazon Security Lake AWS Log Source.
---

# Resource: aws_securitylake_aws_log_source

Terraform resource for managing an Amazon Security Lake AWS Log Source.

## Example Usage

### Basic Usage

```terraform
resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = ["123456789012"]
    regions        = ["eu-west-1"]
    source_name    = "ROUTE53"
    source_version = "1.0"
  }
}
```

## Argument Reference

The following arguments are required:

* `source` - (Required) Specify the natively-supported AWS service to add as a source in Security Lake.

`source` supports the following:

* `accounts` - (Optional) Specify the AWS account information where you want to enable Security Lake.
* `regions` - (Required) Specify the Regions where you want to enable Security Lake.
* `source_name` - (Required) The name for a AWS source. This must be a Regionally unique value. Valid values: `ROUTE53`, `VPC_FLOW`, `SH_FINDINGS`, `CLOUD_TRAIL_MGMT`, `LAMBDA_EXECUTION`, `S3_DATA`.
* `source_version` - (Optional) The version for a AWS source. This must be a Regionally unique value.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS log sources using the source name. For example:

```terraform
import {
  to = aws_securitylake_aws_log_source.example
  id = "ROUTE53"
}
```

Using `terraform import`, import AWS log sources using the source name. For example:

```console
% terraform import aws_securitylake_aws_log_source.example ROUTE53
```
