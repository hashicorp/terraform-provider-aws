---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_findings_filter"
description: |-
  Provides a resource to manage an Amazon Macie Findings Filter.
---

# Resource: aws_macie2_findings_filter

Provides a resource to manage an [Amazon Macie Findings Filter](https://docs.aws.amazon.com/macie/latest/APIReference/findingsfilters-id.html).

## Example Usage

```terraform
resource "aws_macie2_account" "example" {}

resource "aws_macie2_findings_filter" "test" {
  name        = "NAME OF THE FINDINGS FILTER"
  description = "DESCRIPTION"
  position    = 1
  action      = "ARCHIVE"
  finding_criteria {
    criterion {
      field = "region"
      eq    = [data.aws_region.current.region]
    }
  }
  depends_on = [aws_macie2_account.test]
}
```

## Argument Reference

This resource supports the following arguments:

* `action` - (Required) Action to perform on findings that meet the filter criteria (`finding_criteria`). Valid values are `ARCHIVE`, to suppress (automatically archive) the findings, and `NOOP`, to not perform any action on the findings.
* `description` - (Optional) Custom description of the filter. Can contain as many as 512 characters.
* `finding_criteria` - (Required) Criteria to use to filter findings. See [`finding_criteria` Block](#finding_criteria-block) for details.
* `name` - (Optional) Custom name for the filter. Must contain at least 3 characters and can contain as many as 64 characters. If omitted, Terraform assigns a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `position` - (Optional) Position of the filter in the list of saved filters on the Amazon Macie console. This value also determines the order in which the filter is applied to findings, relative to other filters that are also applied to the findings.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `finding_criteria` Block

The `finding_criteria` configuration block supports the following arguments:

* `criterion` - (Optional) Condition that specifies the property, operator, and one or more values to use to filter the results. See [`criterion` Block](#criterion-block) for details.

### `criterion` Block

The `criterion` configuration block supports the following arguments:

* `eq` - (Optional) Value for the property matches (equals) the specified value. If you specify multiple values, Amazon Macie uses OR logic to join the values.
* `eq_exact_match` - (Optional) Value for the property exclusively matches (equals an exact match for) all the specified values. If you specify multiple values, Amazon Macie uses AND logic to join the values.
* `field` - (Required) Name of the field to be evaluated.
* `gt` - (Optional) Value for the property is greater than the specified value.
* `gte` - (Optional) Value for the property is greater than or equal to the specified value.
* `lt` - (Optional) Value for the property is less than the specified value.
* `lte` - (Optional) Value for the property is less than or equal to the specified value.
* `neq` - (Optional) Value for the property doesn't match (doesn't equal) the specified value. If you specify multiple values, Amazon Macie uses OR logic to join the values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Findings Filter.
* `id` - Unique identifier (ID) of the Macie Findings Filter.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `4m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_macie2_findings_filter` using the ID. For example:

```terraform
import {
  to = aws_macie2_findings_filter.example
  id = "abcd1"
}
```

Using `terraform import`, import `aws_macie2_findings_filter` using the ID. For example:

```console
% terraform import aws_macie2_findings_filter.example abcd1
```
