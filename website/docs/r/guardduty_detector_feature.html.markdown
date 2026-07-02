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

resource "aws_guardduty_detector_feature" "s3_protection" {
  detector_id = aws_guardduty_detector.example.id
  name        = "S3_DATA_EVENTS"
  status      = "ENABLED"
}
```

### Extended Threat Detection for EKS

To enable GuardDuty [Extended Threat Detection](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty-extended-threat-detection.html) for EKS, you need at least one of these features enabled: [EKS Protection](https://docs.aws.amazon.com/guardduty/latest/ug/kubernetes-protection.html) or [Runtime Monitoring](https://docs.aws.amazon.com/guardduty/latest/ug/runtime-monitoring-configuration.html). For maximum detection coverage, enabling both is recommended to enhance detection capabilities.

```terraform
resource "aws_guardduty_detector" "example" {
  enable = true
}

resource "aws_guardduty_detector_feature" "eks_protection" {
  detector_id = aws_guardduty_detector.example.id
  name        = "EKS_AUDIT_LOGS"
  status      = "ENABLED"
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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
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

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_guardduty_detector_feature.s3_protection
  identity = {
    detector_id = "12abc34d567e8fa901bc2d34eexample"
    name        = "S3_DATA_EVENTS"
  }
}

resource "aws_guardduty_detector_feature" "s3_protection" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `detector_id` (String) Amazon GuardDuty detector ID.
* `name` (String) Name of the detector feature.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import GuardDuty Detector Features using the `detector_id` and `name` separated by a forward slash (`/`). For example:

```terraform
import {
  to = aws_guardduty_detector_feature.s3_protection
  id = "12abc34d567e8fa901bc2d34eexample/S3_DATA_EVENTS"
}
```

Using `terraform import`, import GuardDuty Detector Features using the `detector_id` and `name` separated by a forward slash (`/`). For example:

```console
% terraform import aws_guardduty_detector_feature.s3_protection 12abc34d567e8fa901bc2d34eexample/S3_DATA_EVENTS
```
