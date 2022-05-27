---
subcategory: "DLM (Data Lifecycle Manager)"
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
data "aws_caller_identity" "current" {}

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
        "AWS": "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
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

### Example Event Based Policy Usage

```
data "aws_caller_identity" "current" {}

resource "aws_dlm_lifecycle_policy" "example" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.example.arn

  policy_details {
    policy_type = "EVENT_BASED_POLICY"

    action {
      name = "tf-acc-basic"
      cross_region_copy {
        encryption_configuration {}
        retain_rule {
          interval      = 15
          interval_unit = "MONTHS"
        }

        target = %[1]q
      }
    }

    event_source {
      type = "MANAGED_CWE"
      parameters {
        description_regex = "^.*Created for policy: policy-1234567890abcdef0.*$"
        event_type        = "shareSnapshot"
        snapshot_owner    = [data.aws_caller_identity.current.account_id]
      }
    }
  }
}

data "aws_iam_policy" "example" {
  name = "AWSDataLifecycleManagerServiceRole"
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.id
  policy_arn = data.aws_iam_policy.example.arn
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

* `action` - (Optional) The actions to be performed when the event-based policy is triggered. You can specify only one action per policy. This parameter is required for event-based policies only. If you are creating a snapshot or AMI policy, omit this parameter. See the [`action` configuration](#action-arguments) block.
* `event_source` - (Optional) The event that triggers the event-based policy. This parameter is required for event-based policies only. If you are creating a snapshot or AMI policy, omit this parameter. See the [`event_source` configuration](#event-source-arguments) block.
* `resource_types` - (Optional) A list of resource types that should be targeted by the lifecycle policy. Valid values are `VOLUME` and `INSTANCE`.
* `resource_locations` - (Optional) The location of the resources to backup. If the source resources are located in an AWS Region, specify `CLOUD`. If the source resources are located on an Outpost in your account, specify `OUTPOST`. If you specify `OUTPOST`, Amazon Data Lifecycle Manager backs up all resources of the specified type with matching target tags across all of the Outposts in your account. Valid values are `CLOUD` and `OUTPOST`.
* `policy_type` - (Optional) The valid target resource types and actions a policy can manage. Specify `EBS_SNAPSHOT_MANAGEMENT` to create a lifecycle policy that manages the lifecycle of Amazon EBS snapshots. Specify `IMAGE_MANAGEMENT` to create a lifecycle policy that manages the lifecycle of EBS-backed AMIs. Specify `EVENT_BASED_POLICY` to create an event-based policy that performs specific actions when a defined event occurs in your AWS account. Default value is `EBS_SNAPSHOT_MANAGEMENT`.
* `parameters` - (Optional) A set of optional parameters for snapshot and AMI lifecycle policies. See the [`parameters` configuration](#parameters-arguments) block.
* `schedule` - (Optional) See the [`schedule` configuration](#schedule-arguments) block.
* `target_tags` (Optional) A map of tag keys and their values. Any resources that match the `resource_types` and are tagged with _any_ of these tags will be targeted.

~> Note: You cannot have overlapping lifecycle policies that share the same `target_tags`. Terraform is unable to detect this at plan time but it will fail during apply.

#### Action arguments

* `cross_region_copy` - (Optional) The rule for copying shared snapshots across Regions. See the [`cross_region_copy` configuration](#action-cross-region-copy-rule-arguments) block.
* `name` - (Optional) A descriptive name for the action.

##### Action Cross Region Copy Rule arguments

* `encryption_configuration` - (Required) The encryption settings for the copied snapshot. See the [`encryption_configuration`](#encryption-configuration-arguments) block. Max of 1 per action.
* `retain_rule` - (Required) Specifies the retention rule for cross-Region snapshot copies. See the [`retain_rule`](#cross-region-copy-rule-retain-rule-arguments) block. Max of 1 per action.
* `target` - (Required) The target Region or the Amazon Resource Name (ARN) of the target Outpost for the snapshot copies.

###### Encryption Configuration arguments

* `cmk_arn` - (Optional) The Amazon Resource Name (ARN) of the AWS KMS key to use for EBS encryption. If this parameter is not specified, the default KMS key for the account is used.
* `encrypted` - (Required) To encrypt a copy of an unencrypted snapshot when encryption by default is not enabled, enable encryption using this parameter. Copies of encrypted snapshots are encrypted, even if this parameter is false or when encryption by default is not enabled.

#### Event Source arguments

* `parameters` - (Required) Information about the event. See the [`parameters` configuration](#event-source-parameters-arguments) block.
* `type` - (Required) The source of the event. Currently only managed CloudWatch Events rules are supported. Valid values are `MANAGED_CWE`.

##### Event Source Parameters arguments

* `description_regex` - (Required) The snapshot description that can trigger the policy. The description pattern is specified using a regular expression. The policy runs only if a snapshot with a description that matches the specified pattern is shared with your account.
* `event_type` - (Required) The type of event. Currently, only `shareSnapshot` events are supported.
* `snapshot_owner` - (Required) The IDs of the AWS accounts that can trigger policy by sharing snapshots with your account. The policy only runs if one of the specified AWS accounts shares a snapshot with your account.

#### Parameters arguments

* `exclude_boot_volume` - (Optional) Indicates whether to exclude the root volume from snapshots created using CreateSnapshots. The default is `false`.
* `no_reboot` - (Optional) Applies to AMI lifecycle policies only. Indicates whether targeted instances are rebooted when the lifecycle policy runs. `true` indicates that targeted instances are not rebooted when the policy runs. `false` indicates that target instances are rebooted when the policy runs. The default is `true` (instances are not rebooted).

#### Schedule arguments

* `copy_tags` - (Optional) Copy all user-defined tags on a source volume to snapshots of the volume created by this policy.
* `create_rule` - (Required) See the [`create_rule`](#create-rule-arguments) block. Max of 1 per schedule.
* `cross_region_copy_rule` (Optional) - See the [`cross_region_copy_rule`](#cross-region-copy-rule-arguments) block. Max of 3 per schedule.
* `name` - (Required) A name for the schedule.
* `deprecate_rule` - (Required) See the [`deprecate_rule`](#deprecate-rule-arguments) block. Max of 1 per schedule.
* `fast_restore_rule` - (Required) See the [`fast_restore_rule`](#fast-restore-rule-arguments) block. Max of 1 per schedule.
* `retain_rule` - (Required) See the [`retain_rule`](#retain-rule-arguments) block. Max of 1 per schedule.
* `share_rule` - (Required) See the [`share_rule`](#share-rule-arguments) block. Max of 1 per schedule.
* `tags_to_add` - (Optional) A map of tag keys and their values. DLM lifecycle policies will already tag the snapshot with the tags on the volume. This configuration adds extra tags on top of these.
* `variable_tags` - (Optional) A map of tag keys and variable values, where the values are determined when the policy is executed. Only `$(instance-id)` or `$(timestamp)` are valid values. Can only be used when `resource_types` is `INSTANCE`.

#### Create Rule arguments

* `cron_expression` - (Optional) The schedule, as a Cron expression. The schedule interval must be between 1 hour and 1 year.
* `interval` - (Optional) How often this lifecycle policy should be evaluated. `1`, `2`,`3`,`4`,`6`,`8`,`12` or `24` are valid values.
* `interval_unit` - (Optional) The unit for how often the lifecycle policy should be evaluated. `HOURS` is currently the only allowed value and also the default value.
* `location` - (Optional) Specifies the destination for snapshots created by the policy. To create snapshots in the same Region as the source resource, specify `CLOUD`. To create snapshots on the same Outpost as the source resource, specify `OUTPOST_LOCAL`. If you omit this parameter, `CLOUD` is used by default. If the policy targets resources in an AWS Region, then you must create snapshots in the same Region as the source resource. If the policy targets resources on an Outpost, then you can create snapshots on the same Outpost as the source resource, or in the Region of that Outpost. Valid values are `CLOUD` and `OUTPOST_LOCAL`.
* `times` - (Optional) A list of times in 24 hour clock format that sets when the lifecycle policy should be evaluated. Max of 1.

#### Deprecate Rule arguments

* `count` - (Optional) Specifies the number of oldest AMIs to deprecate. Must be an integer between `1` and `1000`.
* `interval` - (Optional) Specifies the period after which to deprecate AMIs created by the schedule. The maximum is 100 years. This is equivalent to 1200 months, 5200 weeks, or 36500 days.
* `interval_unit` - (Optional) The unit of time for time-based retention. Valid values are `DAYS`, `WEEKS`, `MONTHS`, `YEARS`.

#### Fast Restore Rule arguments

* `availability_zones` - (Required) The Availability Zones in which to enable fast snapshot restore.
* `count` - (Optional) The number of snapshots to be enabled with fast snapshot restore. Must be an integer between `1` and `1000`.
* `interval` - (Optional) The amount of time to enable fast snapshot restore. The maximum is 100 years. This is equivalent to 1200 months, 5200 weeks, or 36500 days.
* `interval_unit` - (Optional) The unit of time for enabling fast snapshot restore. Valid values are `DAYS`, `WEEKS`, `MONTHS`, `YEARS`.

#### Retain Rule arguments

* `count` - (Optional) How many snapshots to keep. Must be an integer between `1` and `1000`.
* `interval` - (Optional) The amount of time to retain each snapshot. The maximum is 100 years. This is equivalent to 1200 months, 5200 weeks, or 36500 days.
* `interval_unit` - (Optional) The unit of time for time-based retention. Valid values are `DAYS`, `WEEKS`, `MONTHS`, `YEARS`.

#### Share Rule arguments

* `target_accounts` - (Required) The IDs of the AWS accounts with which to share the snapshots.
* `interval` - (Optional) The period after which snapshots that are shared with other AWS accounts are automatically unshared.
* `interval_unit` - (Optional) The unit of time for the automatic unsharing interval. Valid values are `DAYS`, `WEEKS`, `MONTHS`, `YEARS`.

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
