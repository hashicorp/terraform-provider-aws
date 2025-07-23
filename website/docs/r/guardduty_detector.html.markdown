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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `enable` - (Optional) Enable monitoring and feedback reporting. Setting to `false` is equivalent to "suspending" GuardDuty. Defaults to `true`.
* `finding_publishing_frequency` - (Optional) Specifies the frequency of notifications sent for subsequent finding occurrences. If the detector is a GuardDuty member account, the value is determined by the GuardDuty primary account and cannot be modified, otherwise defaults to `SIX_HOURS`. For standalone and GuardDuty primary accounts, it must be configured in Terraform to enable drift detection. Valid values for standalone and primary accounts: `FIFTEEN_MINUTES`, `ONE_HOUR`, `SIX_HOURS`. See [AWS Documentation](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_findings_cloudwatch.html#guardduty_findings_cloudwatch_notification_frequency) for more information.
* `datasources` - (Optional, **Deprecated** use `aws_guardduty_detector_feature` resources instead) Describes which data sources will be enabled for the detector. See [Data Sources](#data-sources) below for more details. [Deprecated](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty-feature-object-api-changes-march2023.html) in favor of [`aws_guardduty_detector_feature` resources](guardduty_detector_feature.html).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Data Sources

The `datasources` block supports the following:

* `s3_logs` - (Optional) Configures [S3 protection](https://docs.aws.amazon.com/guardduty/latest/ug/s3-protection.html).
  See [S3 Logs](#s3-logs) below for more details.
* `kubernetes` - (Optional) Configures [Kubernetes protection](https://docs.aws.amazon.com/guardduty/latest/ug/kubernetes-protection.html).
  See [Kubernetes](#kubernetes) and [Kubernetes Audit Logs](#kubernetes-audit-logs) below for more details.
* `malware_protection` - (Optional) Configures [Malware Protection](https://docs.aws.amazon.com/guardduty/latest/ug/malware-protection.html).
  See [Malware Protection](#malware-protection), [Scan EC2 instance with findings](#scan-ec2-instance-with-findings) and [EBS volumes](#ebs-volumes) below for more details.

The `datasources` block is deprecated since March 2023. Use the `features` block instead and [map each `datasources` block to the corresponding `features` block](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty-feature-object-api-changes-march2023.html#guardduty-feature-enablement-datasource-relation).

### S3 Logs

The `s3_logs` block supports the following:

* `enable` - (Required) If true, enables [S3 protection](https://docs.aws.amazon.com/guardduty/latest/ug/s3-protection.html).
  Defaults to `true`.

### Kubernetes

The `kubernetes` block supports the following:

* `audit_logs` - (Required) Configures Kubernetes audit logs as a data source for [Kubernetes protection](https://docs.aws.amazon.com/guardduty/latest/ug/kubernetes-protection.html).
  See [Kubernetes Audit Logs](#kubernetes-audit-logs) below for more details.

### Kubernetes Audit Logs

The `audit_logs` block supports the following:

* `enable` - (Required) If true, enables Kubernetes audit logs as a data source for [Kubernetes protection](https://docs.aws.amazon.com/guardduty/latest/ug/kubernetes-protection.html).
  Defaults to `true`.

### Malware Protection

`malware_protection` block supports the following:

* `scan_ec2_instance_with_findings` - (Required) Configure whether [Malware Protection](https://docs.aws.amazon.com/guardduty/latest/ug/malware-protection.html) is enabled as data source for EC2 instances with findings for the detector.
  See [Scan EC2 instance with findings](#scan-ec2-instance-with-findings) below for more details.

#### Scan EC2 instance with findings

The `scan_ec2_instance_with_findings` block supports the following:

* `ebs_volumes` - (Required) Configure whether scanning EBS volumes is enabled as data source for the detector for instances with findings.
  See [EBS volumes](#ebs-volumes) below for more details.

#### EBS volumes

The `ebs_volumes` block supports the following:

* `enable` - (Required) If true, enables [Malware Protection](https://docs.aws.amazon.com/guardduty/latest/ug/malware-protection.html) as data source for the detector.
  Defaults to `true`.

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
