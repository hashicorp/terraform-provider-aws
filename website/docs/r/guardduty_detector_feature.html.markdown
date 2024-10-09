---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_detector_feature"
description: |-
  Provides a resource to manage an Amazon GuardDuty detector feature
---

# Resource: aws_guardduty_detector_feature

Provides a resource to manage a single Amazon GuardDuty [detector feature](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty-features-activation-model.html#guardduty-features).

~> **NOTE:** Deleting this resource does not disable the detector feature, the resource in simply removed from state instead.

## Example Usage

```terraform
resource "aws_guardduty_detector" "example" {
  enable = true
}

resource "aws_guardduty_detector_feature" "eks_runtime_monitoring" {
  detector_id = aws_guardduty_detector.example.id
  name        = "EKS_RUNTIME_MONITORING"
  status      = "ENABLED"

  additional_configuration {
    name   = "EKS_ADDON_MANAGEMENT"
    status = "ENABLED"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `detector_id` - (Required) Amazon GuardDuty detector ID.
* `name` - (Required) The name of the detector feature. Valid values: `S3_DATA_EVENTS`, `EKS_AUDIT_LOGS`, `EBS_MALWARE_PROTECTION`, `RDS_LOGIN_EVENTS`, `EKS_RUNTIME_MONITORING`, `LAMBDA_NETWORK_LOGS`, `RUNTIME_MONITORING`. Only one of two features `EKS_RUNTIME_MONITORING` or `RUNTIME_MONITORING` can be added, adding both features will cause an error. Refer to the [AWS Documentation](https://docs.aws.amazon.com/guardduty/latest/APIReference/API_DetectorFeatureConfiguration.html) for the current list of supported values.
* `status` - (Required) The status of the detector feature. Valid values: `ENABLED`, `DISABLED`.
* `additional_configuration` - (Optional) Additional feature configuration block for features`EKS_RUNTIME_MONITORING` or `RUNTIME_MONITORING`. See [below](#additional-configuration).

### Additional Configuration

The `additional_configuration` block supports the following:

* `name` - (Required) The name of the additional configuration for a feature. Valid values: `EKS_ADDON_MANAGEMENT`, `ECS_FARGATE_AGENT_MANAGEMENT`, `EC2_AGENT_MANAGEMENT`. Refer to the [AWS Documentation](https://docs.aws.amazon.com/guardduty/latest/APIReference/API_DetectorAdditionalConfiguration.html) for the current list of supported values.
* `status` - (Required) The status of the additional configuration. Valid values: `ENABLED`, `DISABLED`.

## Attribute Reference

This resource exports no additional attributes.
