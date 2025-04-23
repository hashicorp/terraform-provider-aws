---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_detector"
description: |-
  Provides a resource to manage an Amazon GuardDuty detector
---

# Resource: aws_guardduty_detector

Provides a resource to manage an Amazon GuardDuty detector.

~> **NOTE:** Deleting this resource is equivalent to "disabling" GuardDuty for an AWS region, which removes all existing findings. You can set the `enable` attribute to `false` to instead "suspend" monitoring and feedback reporting while keeping existing data. See the [Suspending or Disabling Amazon GuardDuty documentation](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_suspend-disable.html) for more information.

## Example Usage

```terraform
resource "aws_guardduty_detector" "example" {
  enable = true

}
```

## Argument Reference

This resource supports the following arguments:

* `enable` - (Optional) Enable monitoring and feedback reporting. Setting to `false` is equivalent to "suspending" GuardDuty. Defaults to `true`.
* `finding_publishing_frequency` - (Optional) Specifies the frequency of notifications sent for subsequent finding occurrences. If the detector is a GuardDuty member account, the value is determined by the GuardDuty primary account and cannot be modified, otherwise defaults to `SIX_HOURS`. For standalone and GuardDuty primary accounts, it must be configured in Terraform to enable drift detection. Valid values for standalone and primary accounts: `FIFTEEN_MINUTES`, `ONE_HOUR`, `SIX_HOURS`. See [AWS Documentation](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_findings_cloudwatch.html#guardduty_findings_cloudwatch_notification_frequency) for more information.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `account_id` - The AWS account ID of the GuardDuty detector
* `arn` - Amazon Resource Name (ARN) of the GuardDuty detector
* `id` - The ID of the GuardDuty detector
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import GuardDuty detectors using the detector ID. For example:

```terraform
import {
  to = aws_guardduty_detector.MyDetector
  id = "00b00fd5aecc0ab60a708659477e9617"
}
```

Using `terraform import`, import GuardDuty detectors using the detector ID. For example:

```console
% terraform import aws_guardduty_detector.MyDetector 00b00fd5aecc0ab60a708659477e9617
```

The ID of the detector can be retrieved via the [AWS CLI](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/guardduty/list-detectors.html) using `aws guardduty list-detectors`.
