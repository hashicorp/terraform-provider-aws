---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_detector"
description: |-
  Retrieve information about a GuardDuty detector.
---

# Data Source: aws_guardduty_detector

Retrieve information about a GuardDuty detector.

## Example Usage

```terraform
data "aws_guardduty_detector" "example" {}
```

## Argument Reference

* `id` - (Optional) ID of the detector.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `finding_publishing_frequency` - The frequency of notifications sent about subsequent finding occurrences.
* `service_role_arn` - Service-linked role that grants GuardDuty access to the resources in the AWS account.
* `status` - Current status of the detector.
