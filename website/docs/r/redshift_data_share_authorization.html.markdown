---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_data_share_authorization"
description: |-
  Terraform resource for managing an AWS Redshift Data Share Authorization.
---
# Resource: aws_redshift_data_share_authorization

Terraform resource for managing an AWS Redshift Data Share Authorization.

## Example Usage

### Basic Usage

```terraform
resource "aws_redshift_data_share_authorization" "example" {
  consumer_identifier = "123456789012"
  data_share_arn      = "arn:aws:redshift:us-west-2:123456789012:datashare:3072dae5-022b-4d45-9cd3-01f010aae4b2/example_share"
}
```

## Argument Reference

The following arguments are required:

* `consumer_identifier` - (Required) Identifier of the data consumer that is authorized to access the datashare. This identifier is an AWS account ID or a keyword, such as `ADX`.
* `data_share_arn` - (Required) Amazon Resource Name (ARN) of the datashare that producers are to authorize sharing for.

The following arguments are optional:

* `allow_writes` - (Optional) Whether to allow write operations for a datashare.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A comma-delimited string concatenating `data_share_arn` and `consumer_identifier`.
* `managed_by` - Identifier of a datashare to show its managing entity.
* `producer_arn` - Amazon Resource Name (ARN) of the producer.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Data Share Authorization using the `id`. For example:

```terraform
import {
  to = aws_redshift_data_share_authorization.example
  id = "arn:aws:redshift:us-west-2:123456789012:datashare:3072dae5-022b-4d45-9cd3-01f010aae4b2/example_share,123456789012"
}
```

Using `terraform import`, import Redshift Data Share Authorization using the `id`. For example:

```console
% terraform import aws_redshift_data_share_authorization.example arn:aws:redshift:us-west-2:123456789012:datashare:3072dae5-022b-4d45-9cd3-01f010aae4b2/example_share,123456789012
```
