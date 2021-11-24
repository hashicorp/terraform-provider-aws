---
subcategory: "Data Lifecycle Manager (DLM)"
layout: "aws"
page_title: "AWS: aws_dlm_lifecycle_policy"
description: |-
  Provides a Data Lifecycle Manager (DLM) lifecycle policy for managing snapshots.
---

# Resource: aws_dlm_lifecycle_policy

Provides a [Data Lifecycle Manager (DLM) lifecycle policy](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snapshot-lifecycle.html) for managing snapshots.

## Example Usage

### Basic

```terraform
resource "aws_iam_role" "dlm_lifecycle_role" {
  name = "dlm-lifecycle-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "dlm_lifecycle" {
  name = "dlm-lifecycle-policy"
  role = aws_iam_role.dlm_lifecycle_role.id

  policy = <<EOF
{
   "Version": "2012-10-17",
   "Statement": [
      {
         "Effect": "Allow",
         "Action": [
            "ec2:CreateSnapshot",
            "ec2:CreateSnapshots",
            "ec2:DeleteSnapshot",
            "ec2:DescribeInstances",
            "ec2:DescribeVolumes",
            "ec2:DescribeSnapshots"
         ],
         "Resource": "*"
      },
      {
         "Effect": "Allow",
         "Action": [
            "ec2:CreateTags"
         ],
         "Resource": "arn:aws:ec2:*::snapshot/*"
      }
   ]
}
EOF
}

resource "aws_dlm_lifecycle_policy" "example" {
  description        = "example DLM lifecycle policy"
  execution_role_arn = aws_iam_role.dlm_lifecycle_role.arn
  state              = "ENABLED"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "2 weeks of daily snapshots"

      create_rule {
        interval      = 24
        interval_unit = "HOURS"
        times         = ["23:45"]
      }

      retain_rule {
        count = 14
      }

      tags_to_add = {
        SnapshotCreator = "DLM"
      }

      copy_tags = false
    }

    target_tags = {
      Snapshot = "true"
    }
  }
}
```

### Example Cross-Region Snapshot Copy Usage

```terraform
# ...other configuration...
resource "aws_kms_key" "dlm_cross_region_copy_cmk" {
  provider = aws.alternate

  description = "Example Alternate Region KMS Key"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "dlm-cross-region-copy-cmk",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_dlm_lifecycle_policy" "example" {
  description        = "example DLM lifecycle policy"
  execution_role_arn = aws_iam_role.dlm_lifecycle_role.arn
  state              = "ENABLED"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "2 weeks of daily snapshots"

      create_rule {
        interval      = 24
        interval_unit = "HOURS"
        times         = ["23:45"]
      }

      retain_rule {
        count = 14
      }

      tags_to_add = {
        SnapshotCreator = "DLM"
      }

      copy_tags = false

      cross_region_copy_rule {
        target    = "us-west-2"
        encrypted = true
        cmk_arn   = aws_kms_key.dlm_cross_region_copy_cmk.arn
        copy_tags = true
        retain_rule {
          interval      = 30
          interval_unit = "DAYS"
        }
      }
    }

    target_tags = {
      Snapshot = "true"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Required) A description for the DLM lifecycle policy.
* `execution_role_arn` - (Required) The ARN of an IAM role that is able to be assumed by the DLM service.
* `policy_details` - (Required) See the [`policy_details` configuration](#policy-details-arguments) block. Max of 1.
* `state` - (Optional) Whether the lifecycle policy should be enabled or disabled. `ENABLED` or `DISABLED` are valid values. Defaults to `ENABLED`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

#### Policy Details arguments

* `resource_types` - (Required) A list of resource types that should be targeted by the lifecycle policy. `VOLUME` is currently the only allowed value.
* `schedule` - (Required) See the [`schedule` configuration](#schedule-arguments) block.
* `target_tags` (Required) A map of tag keys and their values. Any resources that match the `resource_types` and are tagged with _any_ of these tags will be targeted.

~> Note: You cannot have overlapping lifecycle policies that share the same `target_tags`. Terraform is unable to detect this at plan time but it will fail during apply.

#### Schedule arguments

* `copy_tags` - (Optional) Copy all user-defined tags on a source volume to snapshots of the volume created by this policy.
* `create_rule` - (Required) See the [`create_rule`](#create-rule-arguments) block. Max of 1 per schedule.
* `cross_region_copy_rule` (Optional) - See the [`cross_region_copy_rule`](#cross-region-copy-rule-arguments) block. Max of 3 per schedule.
* `name` - (Required) A name for the schedule.
* `retain_rule` - (Required) See the [`retain_rule`](#retain-rule-arguments) block. Max of 1 per schedule.
* `tags_to_add` - (Optional) A map of tag keys and their values. DLM lifecycle policies will already tag the snapshot with the tags on the volume. This configuration adds extra tags on top of these.

#### Create Rule arguments

* `interval` - (Required) How often this lifecycle policy should be evaluated. `1`, `2`,`3`,`4`,`6`,`8`,`12` or `24` are valid values.
* `interval_unit` - (Optional) The unit for how often the lifecycle policy should be evaluated. `HOURS` is currently the only allowed value and also the default value.
* `times` - (Optional) A list of times in 24 hour clock format that sets when the lifecycle policy should be evaluated. Max of 1.

#### Retain Rule arguments

* `count` - (Required) How many snapshots to keep. Must be an integer between 1 and 1000.

#### Cross Region Copy Rule arguments

* `cmk_arn` - (Optional) The Amazon Resource Name (ARN) of the AWS KMS customer master key (CMK) to use for EBS encryption. If this argument is not specified, the default KMS key for the account is used.
* `copy_tags` - (Optional) Whether to copy all user-defined tags from the source snapshot to the cross-region snapshot copy.
* `deprecate_rule` - (Optional) The AMI deprecation rule for cross-Region AMI copies created by the rule. See the [`deprecate_rule`](#cross-region-copy-rule-deprecate-rule-arguments) block.
* `encrypted` - (Required) To encrypt a copy of an unencrypted snapshot if encryption by default is not enabled, enable encryption using this parameter. Copies of encrypted snapshots are encrypted, even if this parameter is false or if encryption by default is not enabled.
* `retain_rule` - (Required) The retention rule that indicates how long snapshot copies are to be retained in the destination Region. See the [`retain_rule`](#cross-region-copy-rule-retain-rule-arguments) block. Max of 1 per schedule.
* `target` - (Required) The target Region or the Amazon Resource Name (ARN) of the target Outpost for the snapshot copies.

#### Cross Region Copy Rule Deprecate Rule arguments

* `interval` - (Required) The period after which to deprecate the cross-Region AMI copies. The period must be less than or equal to the cross-Region AMI copy retention period, and it can't be greater than 10 years. This is equivalent to 120 months, 520 weeks, or 3650 days.
* `interval_unit` - (Required) The unit of time in which to measure the `interval`. Valid values: `DAYS`, `WEEKS`, `MONTHS`, or `YEARS`.

#### Cross Region Copy Rule Retain Rule arguments

* `interval` - (Required) The amount of time to retain each snapshot. The maximum is 100 years. This is equivalent to 1200 months, 5200 weeks, or 36500 days.
* `interval_unit` - (Required) The unit of time for time-based retention. Valid values: `DAYS`, `WEEKS`, `MONTHS`, or `YEARS`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the DLM Lifecycle Policy.
* `id` - Identifier of the DLM Lifecycle Policy.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

DLM lifecycle policies can be imported by their policy ID:

```
$ terraform import aws_dlm_lifecycle_policy.example policy-abcdef12345678901
```
