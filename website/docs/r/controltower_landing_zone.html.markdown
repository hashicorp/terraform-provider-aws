---
subcategory: "Control Tower"
layout: "aws"
page_title: "AWS: aws_controltower_landing_zone"
description: |-
  Creates a new landing zone using Control Tower.
---

# Resource: aws_controltower_landing_zone

Creates a new landing zone using Control Tower. For more information on usage, please see the
[AWS Control Tower Landing Zone User Guide](https://docs.aws.amazon.com/controltower/latest/userguide/how-control-tower-works.html).

## Example Usage

```terraform
resource "aws_controltower_landing_zone" "example" {
  manifest {
    governed_regions = ["us-west-2"]

    centralized_logging {
      configurations {
        access_logging_bucket {
          retention_days = 3650
        }
        logging_bucket {
          retention_days = 365
        }
      }
    }

    organization_structure {
      sandbox {
        name = "Sandbox"
      }
      security {
        name = "Security"
      }
    }
  }

  version = "3.2"
}
```

## Argument Reference

This resource supports the following arguments:

* `manifest` - (Required) Landing zone configuration. For examples, review [Launch your landing zone](https://docs.aws.amazon.com/controltower/latest/userguide/lz-api-launch).
  * `access_management` - (Optional) Access management configuration.
    * `enabled` - (Optional) Whether AWS Control Tower sets up AWS account access with AWS Identity and Access Management (IAM), or whether to self-manage AWS account access.
  * `centralized_logging` - (Optional) Log configuration for Amazon S3.
      * `account_id` - (Optional) The AWS account ID for centralized logging.
      * `configurations` - (Optional) Configurations.
        * `access_logging_bucket` - (Optional) Amazon S3 bucket retention for access logging.
          * `retention_days` - (Optional) Retention period for access logging bucket.
        * `kms_key_arn` - (Optional) KMS key ARN used by CloudTrail and Config service to encrypt data in logging bucket.
        * `logging_bucket` - (Optional) Amazon S3 bucket retention for logging.
          * `retention_days` - (Optional) Retention period for centralized logging bucket.
      * `enabled` - (Optional) Whether or not logging is enabled.
  * `governed_regions` - (Required) AWS Regions to govern.
  * `organization_structure` - (Optional) Organization structure.
    * `sandbox` - (Optional) Sandbox Organizational Unit configuration.
      * `name` - (Optional) The sandbox Organizational Unit name.
    * `security` - (Optional) Security Organizational Unit configuration.
      * `name` - (Optional) The security Organizational Unit name.
  * `security_roles` - (Optional) Organization structure.
    * `account_id` - (Optional) The AWS account ID for security roles.
* `tags` - (Optional) Tags to apply to the landing zone. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `version` - (Required) The landing zone version.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The identifier of the landing zone.
* `arn` - The ARN of the landing zone.
* `drift_status` - The drift status summary of the landing zone.
  * `status` - The drift status of the landing zone.
* `latest_available_version` - The latest available version of the landing zone.
* `tags_all` - A map of tags assigned to the landing zone, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Control Tower Landing Zone using the `id`. For example:

```terraform
import {
  to = aws_controltower_landing_zone.example
  id = "1A2B3C4D5E6F7G8H"
}
```

Using `terraform import`, import a Control Tower Landing Zone using the `id`. For example:

```console
% terraform import aws_controltower_landing_zone.example 1A2B3C4D5E6F7G8H
```
