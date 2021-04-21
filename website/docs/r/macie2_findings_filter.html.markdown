---
subcategory: "Macie2"
layout: "aws"
page_title: "AWS: aws_macie2_findings_filter"
description: |-
  Provides a resource to manage an AWS Macie2 Findings Filter.
---

# Resource: aws_macie2_findings_filter

Provides a resource to manage an [AWS Macie Findings Filter](https://docs.aws.amazon.com/macie/latest/APIReference/findingsfilters-id.html).

## Example Usage

```terraform
resource "aws_macie2_account" "example" {
  client_token = "CLIENT TOKEN"
}

resource "aws_macie2_findings_filter" "test" {
  name = "NAME OF THE FINDINGS FILTER"
  client_token = aws_macie2_account.test.client_token
  description = "DESCRIPTION"
  position = 1
  action = "ARCHIVE"
  finding_criteria {
    criterion {
      field  = "region"
      eq = [data.aws_region.current.name]
    }
  }
  depends_on = [aws_macie2_account.test]
}
```

## Argument Reference

The following arguments are supported:

* `finding_criteria` - (Required) The criteria to use to filter findings.
* `client_token` - (Optional) A unique, case-sensitive token that you provide to ensure the idempotency of the request.
* `name` - (Required) A custom name for the filter. The name must contain at least 3 characters and can contain as many as 64 characters.
* `description` - (Optional) A custom description of the filter. The description can contain as many as 512 characters.
* `action` - (Required) The action to perform on findings that meet the filter criteria (`finding_criteria`). Valid values are: `ARCHIVE`, suppress (automatically archive) the findings; and, `NOOP`, don't perform any action on the findings.
* `position` - (Optional) The position of the filter in the list of saved filters on the Amazon Macie console. This value also determines the order in which the filter is applied to findings, relative to other filters that are also applied to the findings.
* `tags` - (Optional) A map of key-value pairs that specifies the tags to associate with the filter.

The `finding_criteria` object supports the following:

* `criterion` -  (Optional) A condition that specifies the property, operator, and one or more values to use to filter the results.  (documented below)

The `criterion` object supports the following:

* `field` - (Required) The name of the field to be evaluated. 
* `eq_exact_match` - (Optional) The value for the property exclusively matches (equals an exact match for) all the specified values. If you specify multiple values, Amazon Macie uses AND logic to join the values.
* `eq` - (Optional) The value for the property matches (equals) the specified value. If you specify multiple values, Macie uses OR logic to join the values.
* `neq` - (Optional) The value for the property doesn't match (doesn't equal) the specified value. If you specify multiple values, Macie uses OR logic to join the values.
* `lt` - (Optional) The value for the property is less than the specified value.
* `lte` - (Optional) The value for the property is less than or equal to the specified value.
* `gt` - (Optional) The value for the property is greater than the specified value.
* `gte` - (Optional) The value for the property is greater than or equal to the specified value.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) of the macie2 Findings Filter.
* `arn` - The Amazon Resource Name (ARN) of the Findings Filter.

## Import

`aws_macie2_findings_filter` can be imported using the id, e.g.

```
$ terraform import aws_macie2_findings_filter.example abcd1
```
