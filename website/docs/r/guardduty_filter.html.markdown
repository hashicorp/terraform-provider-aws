---
layout: "aws"
page_title: "AWS: aws_guardduty_filter"
sidebar_current: "docs-aws-resource-guardduty-filter"
description: |-
  Provides a resource to manage a GuardDuty filter
---

# Resource: aws_guardduty_filter

Provides a resource to manage a GuardDuty filter.

## Example Usage

```hcl
resource "aws_guardduty_filter" "MyFilter" {
  name        = "MyFilter"
  action      = "ARCHIVE"
  description = "Some description"
  detector_id = "123456271278c0df5e089123480d8765"
  rank        = 1

  finding_criteria {
    criterion {
      field     = "region"
      values    = ["eu-west-1"]
      condition = "equals"
    }

    criterion {
      field     = "service.additionalInfo.threatListName"
      values    = ["some-threat"]
      condition = "not_equals"
    }

    criterion {
      field     = "updatedAt"
      values    = ["1570744740000"]
      condition = "less_than"
    }

    criterion {
      field     = "updatedAt"
      values    = ["1570744240000"]
      condition = "greater_than"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `detector_id` - (Required) ID of a GuardDuty detector, attached to your account.
* `name` - (Required) The name of your filter.
* `description` - (Optional) Description of the filter.
* `rank` - (Required) Specifies the position of the filter in the list of current filters. Also specifies the order in which this filter is applied to the findings.
* `action` - (Required) Specifies the action that is to be applied to the findings that match the filter. Can be one of (ARCHIVE|NOOP).
* `tags` (Optional) - The tags that you want to add to the Filter resource. A tag consists of a key and a value.
* `finding_criteria` (Required) - Represents the criteria to be used in the filter for querying findings. A list, consists of `criterion` structures, each having required `condition`, `field` and `values` fields. See the example for the structure and see the [AWS Documentation](https://docs.aws.amazon.com/guardduty/latest/ug/create-filter.html) for the list of available fields for the `field` attribute.

## Attributes Reference

In addition to all arguments above, the following attribute is exported:

* `id` - A compound field, consisting of the ID of the GuardDuty detector and the name of the filter.

## Import

GuardDuty filters can be imported using the detector ID and filter's name separated by underscore, e.g.

```
$ terraform import aws_guardduty_filter.MyFilter 00b00fd5aecc0ab60a708659477e9617_MyFilter
```
