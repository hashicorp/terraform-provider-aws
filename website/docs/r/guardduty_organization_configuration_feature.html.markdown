---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_organization_configuration_feature"
description: |-
  Provides a resource to manage an Amazon GuardDuty organization configuration feature
---

# Resource: aws_guardduty_organization_configuration_feature

Provides a resource to manage a single Amazon GuardDuty [organization configuration feature](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty-features-activation-model.html#guardduty-features).

~> **NOTE:** Deleting this resource does not disable the organization configuration feature, the resource in simply removed from state instead.

## Example Usage

```terraform
resource "aws_guardduty_detector" "example" {
  enable = true
}

resource "aws_guardduty_organization_configuration_feature" "eks_runtime_monitoring" {
  detector_id = aws_guardduty_detector.example.id
  name        = "EKS_RUNTIME_MONITORING"
  auto_enable = "ALL"

  additional_configuration {
    name        = "EKS_ADDON_MANAGEMENT"
    auto_enable = "NEW"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `auto_enable` - (Required) The status of the feature that is configured for the member accounts within the organization. Valid values: `NEW`, `ALL`, `NONE`.
* `detector_id` - (Required) The ID of the detector that configures the delegated administrator.
* `name` - (Required) The name of the feature that will be configured for the organization. Valid values: `S3_DATA_EVENTS`, `EKS_AUDIT_LOGS`, `EBS_MALWARE_PROTECTION`, `RDS_LOGIN_EVENTS`, `EKS_RUNTIME_MONITORING`, `LAMBDA_NETWORK_LOGS`, `RUNTIME_MONITORING`. Only one of two features `EKS_RUNTIME_MONITORING` or `RUNTIME_MONITORING` can be added, adding both features will cause an error. Refer to the [AWS Documentation](https://docs.aws.amazon.com/guardduty/latest/APIReference/API_DetectorFeatureConfiguration.html) for the current list of supported values.
* `additional_configuration` - (Optional) Additional feature configuration block for features `EKS_RUNTIME_MONITORING` or `RUNTIME_MONITORING`. See [below](#additional-configuration).

### Additional Configuration

The `additional_configuration` block supports the following:

* `auto_enable` - (Required) The status of the additional configuration that will be configured for the organization. Valid values: `NEW`, `ALL`, `NONE`.
* `name` - (Required) The name of the additional configuration for a feature that will be configured for the organization. Valid values: `EKS_ADDON_MANAGEMENT`, `ECS_FARGATE_AGENT_MANAGEMENT`, `EC2_AGENT_MANAGEMENT`. Refer to the [AWS Documentation](https://docs.aws.amazon.com/guardduty/latest/APIReference/API_DetectorAdditionalConfiguration.html) for the current list of supported values.

## Attribute Reference

This resource exports no additional attributes.
