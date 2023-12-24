---
subcategory: "Security Lake"
layout: "aws"
page_title: "AWS: aws_securitylake_log_source"
description: |-
  Terraform resource for managing an AWS Security Lake Aws Log Source.
---

# Resource: aws_securitylake_log_source

Terraform resource for managing an AWS Security Lake Aws Log Source.

## Example Usage

### Basic Usage

```terraform
resource "aws_securitylake_log_source" "example" {
	sources {
    accounts 	     = ["1234567890"] 
		regions        = ["eu-west-1"]
		source_name    = "ROUTE53"
		source_version = "1.0"
	}
	depends_on = [aws_securitylake_data_lake.example]
}
```

## Argument Reference

The following arguments are required:

* `sources` - (Required) Specify the natively-supported AWS service to add as a source in Security Lake.

Sources support the following:

* `accounts` - (Optional) Specify the AWS account information where you want to enable Security Lake.
* `regions` - (Required) Specify the Regions where you want to enable Security Lake.
* `source_name` - (Required) The name for a AWS source. This must be a Regionally unique value.Valid Values: ROUTE53 | VPC_FLOW | SH_FINDINGS | CLOUD_TRAIL_MGMT | LAMBDA_EXECUTION | S3_DATA
* `source_version` - (Required) The version for a AWS source. This must be a Regionally unique value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub standards subscriptions using the standards subscription ARN. For example:

```terraform
import {
  to = aws_securitylake_log_source.example
  id = "ROUTE53/1.0"
}
```

Using `terraform import`, import Security Hub standards subscriptions using the standards subscription ARN. For example:

```console
% terraform import aws_securitylake_log_source.example ROUTE53/1.0
```
