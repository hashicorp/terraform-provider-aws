---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_data_share_consumer_association"
description: |-
  Terraform resource for managing an AWS Redshift Data Share Consumer Association.
---
# Resource: aws_redshift_data_share_consumer_association

Terraform resource for managing an AWS Redshift Data Share Consumer Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_redshift_data_share_consumer_association" "example" {
  data_share_arn           = "arn:aws:redshift:us-west-2:012345678901:datashare:b3bfde75-73fd-408b-9086-d6fccfd6d588/example"
  associate_entire_account = true
}
```

### Consumer Region

```terraform
resource "aws_redshift_data_share_consumer_association" "example" {
  data_share_arn  = "arn:aws:redshift:us-west-2:012345678901:datashare:b3bfde75-73fd-408b-9086-d6fccfd6d588/example"
  consumer_region = "us-west-2"
}
```

## Argument Reference

The following arguments are required:

* `data_share_arn` - (Required) Amazon Resource Name (ARN) of the datashare that the consumer is to use with the account or the namespace.

The following arguments are optional:

* `allow_writes` - (Optional) Whether to allow write operations for a datashare.
* `associate_entire_account` - (Optional) Whether the datashare is associated with the entire account. Conflicts with `consumer_arn` and `consumer_region`.
* `consumer_arn` - (Optional) Amazon Resource Name (ARN) of the consumer that is associated with the datashare. Conflicts with `associate_entire_account` and `consumer_region`.
* `consumer_region` - (Optional) From a datashare consumer account, associates a datashare with all existing and future namespaces in the specified AWS Region. Conflicts with `associate_entire_account` and `consumer_arn`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A comma-delimited string concatenating `data_share_arn` and `associate_entire_account`, `consumer_arn`, and `consumer_region`. As only one of the final three arguments can be specified, the other two will always be empty.
* `managed_by` - Identifier of a datashare to show its managing entity.
* `producer_arn` - Amazon Resource Name (ARN) of the producer.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Data Share Consumer Association using the `id`. For example:

```terraform
import {
  to = aws_redshift_data_share_consumer_association.example
  id = "arn:aws:redshift:us-west-2:012345678901:datashare:b3bfde75-73fd-408b-9086-d6fccfd6d588/example,,,us-west-2"
}
```

Using `terraform import`, import Redshift Data Share Consumer Association using the `id`. For example:

```console
% terraform import aws_redshift_data_share_consumer_association.example arn:aws:redshift:us-west-2:012345678901:datashare:b3bfde75-73fd-408b-9086-d6fccfd6d588/example,,,us-west-2
```
