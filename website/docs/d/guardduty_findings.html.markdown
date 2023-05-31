---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_findings"
description: |-
  Terraform data source for managing an AWS GuardDuty Findings.
---

# Data Source: aws_guardduty_findings

Terraform data source for managing an AWS GuardDuty Findings.

## Example Usage

### Basic Usage

```terraform
data "aws_guardduty_findings" "example" {
  detector_id = aws_guardduty_detector.example.id
}
```

## Argument Reference

The following arguments are required:

* `detector_id` - (Required) ID of the GuardDuty detector.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `finding_ids` - A list of finding IDs for the specified detector.
