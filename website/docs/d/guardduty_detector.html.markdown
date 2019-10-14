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
* `status` - The current status of the detector.
* `service_role_arn` - The service-linked role that grants GuardDuty access to the resources in the AWS account.
* `finding_publishing_frequency` - The frequency of notifications sent about subsequent finding occurrences.
