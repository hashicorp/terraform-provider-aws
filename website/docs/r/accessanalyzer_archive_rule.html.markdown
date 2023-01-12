---
subcategory: "IAM Access Analyzer"
layout: "aws"
page_title: "AWS: aws_accessanalyzer_archive_rule"
description: |-
  Terraform resource for managing an AWS AccessAnalyzer Archive Rule.
---

# Resource: aws_accessanalyzer_archive_rule

Terraform resource for managing an AWS AccessAnalyzer Archive Rule.

## Example Usage

### Basic Usage

```terraform
resource "aws_accessanalyzer_archive_rule" "example" {
  analyzer_name = "example-analyzer"
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
* `filter` - (Required) Filter criteria for the archive rule. See [Filter](#filter) for more details.
* `rule_name` - (Required) Rule name.

### Filter

**Note** One comparator must be included with each filter.

* `criteria` - (Required) Filter criteria.
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
$ terraform import aws_accessanalyzer_archive_rule.example example-analyzer/example-rule
```
