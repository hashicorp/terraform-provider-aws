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
resource "aws_guardduty_detector" "MyDetector" {
  enable = true

  datasources {
    s3_logs {
      enable = true
    }
    kubernetes {
      audit_logs {
        enable = false
      }
    }
    malware_protection {
      scan_ec2_instance_with_findings {
        ebs_volumes {
          enable = true
        }
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `datasources` - (Optional, **Deprecated** use `aws_guardduty_detector_feature` resources instead) Data sources to enable for the detector. See [`datasources`](#datasources-block) below for more details. [Deprecated](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty-feature-object-api-changes-march2023.html) in favor of [`aws_guardduty_detector_feature` resources](guardduty_detector_feature.html).
* `enable` - (Optional) Enable monitoring and feedback reporting. Setting to `false` is equivalent to "suspending" GuardDuty. Defaults to `true`.
* `finding_publishing_frequency` - (Optional) Frequency of notifications sent for subsequent finding occurrences. If the detector is a GuardDuty member account, the value is determined by the GuardDuty primary account and cannot be modified, otherwise defaults to `SIX_HOURS`. For standalone and GuardDuty primary accounts, it must be configured in Terraform to enable drift detection. Valid values for standalone and primary accounts: `FIFTEEN_MINUTES`, `ONE_HOUR`, `SIX_HOURS`. See [AWS Documentation](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_findings_cloudwatch.html#guardduty_findings_cloudwatch_notification_frequency) for more information.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `datasources` Block

The `datasources` block is deprecated since March 2023. Use the `features` block instead and [map each `datasources` block to the corresponding `features` block](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty-feature-object-api-changes-march2023.html#guardduty-feature-enablement-datasource-relation).

The `datasources` block supports the following:

* `kubernetes` - (Optional) Configures [Kubernetes protection](https://docs.aws.amazon.com/guardduty/latest/ug/kubernetes-protection.html). See [`kubernetes`](#kubernetes-block) below for more details.
* `malware_protection` - (Optional) Configures [Malware Protection](https://docs.aws.amazon.com/guardduty/latest/ug/malware-protection.html). See [`malware_protection`](#malware_protection-block) below for more details.
* `s3_logs` - (Optional) Configures [S3 protection](https://docs.aws.amazon.com/guardduty/latest/ug/s3-protection.html). See [`s3_logs`](#s3_logs-block) below for more details.

### `s3_logs` Block

The `s3_logs` block supports the following:

* `enable` - (Required) If true, enables [S3 protection](https://docs.aws.amazon.com/guardduty/latest/ug/s3-protection.html). Defaults to `true`.

### `kubernetes` Block

The `kubernetes` block supports the following:

* `audit_logs` - (Required) Configures Kubernetes audit logs as a data source for [Kubernetes protection](https://docs.aws.amazon.com/guardduty/latest/ug/kubernetes-protection.html). See [`audit_logs`](#audit_logs-block) below for more details.

### `audit_logs` Block

The `audit_logs` block supports the following:

* `enable` - (Required) If true, enables Kubernetes audit logs as a data source for [Kubernetes protection](https://docs.aws.amazon.com/guardduty/latest/ug/kubernetes-protection.html). Defaults to `true`.

### `malware_protection` Block

The `malware_protection` block supports the following:

* `scan_ec2_instance_with_findings` - (Required) Configure whether [Malware Protection](https://docs.aws.amazon.com/guardduty/latest/ug/malware-protection.html) is enabled as data source for EC2 instances with findings for the detector. See [`scan_ec2_instance_with_findings`](#scan_ec2_instance_with_findings-block) below for more details.

### `scan_ec2_instance_with_findings` Block

The `scan_ec2_instance_with_findings` block supports the following:

* `ebs_volumes` - (Required) Configure whether scanning EBS volumes is enabled as data source for the detector for instances with findings. See [`ebs_volumes`](#ebs_volumes-block) below for more details.

### `ebs_volumes` Block

The `ebs_volumes` block supports the following:

* `enable` - (Required) If true, enables [Malware Protection](https://docs.aws.amazon.com/guardduty/latest/ug/malware-protection.html) as data source for the detector. Defaults to `true`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `account_id` - AWS account ID of the GuardDuty detector
* `arn` - ARN of the GuardDuty detector
* `id` - ID of the GuardDuty detector
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
