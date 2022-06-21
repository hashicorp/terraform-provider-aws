---
subcategory: "AccessAnalyzer"
layout: "aws"
page_title: "AWS: aws_accessanalyzer_archiverule"
description: |-
  Terraform resource for managing an AWS AccessAnalyzer ArchiveRule.
---

# Resource: aws_accessanalyzer_archiverule

Terraform resource for managing an AWS AccessAnalyzer ArchiveRule.

## Example Usage

### Basic Usage

```terraform
resource "aws_accessanalyzer_archiverule" "example" {
  analyser_name = "example-analyzer"
  rule_name     = "example-rule"

  filter {
    criteria = "condition.aws:UserId"
    eq       = ["userid"]
  }

  filter {
    criteria = "error"
    exists   = true
  }

  filter {
    criteria = "isPublic"
    eq       = ["false"]
  }
}
```

## Argument Reference

The following arguments are required:

* `analyzer_name` - (Required) Analyzer name.
* `filter` - (Required) The filter criteria for the archive rule. See [Filter](#filter) for more details.
* `rule_name` - (Required) Rule name.

### Filter

**Note** At least one comparator must be included with each filter.

* `criteria` - (Required) The filter criteria.
* `contains` - (Optional) Contains comparator.
* `eq` - (Optional) Equals comparator.
* `exists` - (Optional) Boolean comparator.
* `neq` - (Optional) Not Equals comparator.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Resource ID in the format: `analyzer_name/rule_name`.

## Import

AccessAnalyzer ArchiveRule can be imported using the `analyzer_name/rule_name`, e.g.,

```
$ terraform import aws_accessanalyzer_archiverule.example example-analyzer/example-rule
```
