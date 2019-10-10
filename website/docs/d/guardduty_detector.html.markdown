---
layout: "aws"
page_title: "AWS: aws_guardduty_detector"
description: |-
  Retrieve information about a GuardDuty detector.
---

# Data Source: aws_guardduty_detector

Retrieve information about a GuardDuty detector.

## Example Usage

```hcl
data "aws_guardduty_detector" "example" {
}
```

## Argument Reference

* `id` - (optional) The ID of the detector.

## Attributes Reference

* `id` - The ID of the detector.
* `enabled` - State of the detector.
* `service_role_arn` - Service role used by the detector.
* `finding_publishing_frequency` - Finding publishing frequence configured.
