---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_aggregator_v2"
description: |-
  Manages a Security Hub V2 cross-region finding aggregator.
---

# Resource: aws_securityhub_aggregator_v2

Manages a Security Hub V2 Aggregator, which enables cross-region finding aggregation.

~> **NOTE:** Security Hub V2 must be enabled (`aws_securityhub_account_v2`) before creating an aggregator.

## Example Usage

### All Regions

```terraform
resource "aws_securityhub_account_v2" "example" {}

resource "aws_securityhub_aggregator_v2" "example" {
  region_linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_account_v2.example]
}
```

### Specified Regions

```terraform
resource "aws_securityhub_account_v2" "example" {}

resource "aws_securityhub_aggregator_v2" "example" {
  region_linking_mode = "SPECIFIED_REGIONS"
  linked_regions      = ["us-west-2", "eu-west-1"]

  depends_on = [aws_securityhub_account_v2.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `region_linking_mode` - (Required) Determines how Regions are linked to the aggregator. Valid values: `ALL_REGIONS`, `ALL_REGIONS_EXCEPT_SPECIFIED`, `SPECIFIED_REGIONS`.
* `linked_regions` - (Optional) List of Regions linked to the aggregation Region. Required when `region_linking_mode` is `SPECIFIED_REGIONS` or `ALL_REGIONS_EXCEPT_SPECIFIED`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Security Hub V2 Aggregator.
* `aggregation_region` - The AWS Region where data is aggregated.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_securityhub_aggregator_v2.example
  identity = {
    arn = "arn:aws:securityhub:us-east-1:123456789012:aggregator/v2/example"
  }
}

resource "aws_securityhub_aggregator_v2" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the Security Hub V2 aggregator.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub V2 aggregators using `arn`. For example:

```terraform
import {
  to = aws_securityhub_aggregator_v2.example
  id = "arn:aws:securityhub:us-east-1:123456789012:aggregator/v2/example"
}
```

Using `terraform import`, import Security Hub V2 aggregators using `arn`. For example:

```console
% terraform import aws_securityhub_aggregator_v2.example arn:aws:securityhub:us-east-1:123456789012:aggregator/v2/example
```
