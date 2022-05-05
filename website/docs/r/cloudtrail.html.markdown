---
subcategory: "CloudTrail"
layout: "aws"
page_title: "AWS: aws_cloudtrail"
description: |-
  Provides a CloudTrail resource.
---

# Resource: aws_cloudtrail

Provides a CloudTrail resource.

-> **Tip:** For a multi-region trail, this resource must be in the home region of the trail.

-> **Tip:** For an organization trail, this resource must be in the master account of the organization.

## Example Usage

### Basic

Enable CloudTrail to capture all compatible management events in region.
For capturing events from services like IAM, `include_global_service_events` must be enabled.

```terraform
data "aws_caller_identity" "current" {}

resource "aws_cloudtrail" "foobar" {
  name                          = "tf-trail-foobar"
  s3_bucket_name                = aws_s3_bucket.foo.id
  s3_key_prefix                 = "prefix"
  include_global_service_events = false
}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail"
  force_destroy = true
}

resource "aws_s3_bucket_policy" "foo" {
  bucket = aws_s3_bucket.foo.id
  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AWSCloudTrailAclCheck",
            "Effect": "Allow",
            "Principal": {
              "Service": "cloudtrail.amazonaws.com"
            },
            "Action": "s3:GetBucketAcl",
            "Resource": "${aws_s3_bucket.foo.arn}"
        },
        {
            "Sid": "AWSCloudTrailWrite",
            "Effect": "Allow",
            "Principal": {
              "Service": "cloudtrail.amazonaws.com"
            },
            "Action": "s3:PutObject",
            "Resource": "${aws_s3_bucket.foo.arn}/prefix/AWSLogs/${data.aws_caller_identity.current.account_id}/*",
            "Condition": {
                "StringEquals": {
                    "s3:x-amz-acl": "bucket-owner-full-control"
                }
            }
        }
    ]
}
POLICY
}
```

### Data Event Logging

CloudTrail can log [Data Events](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/logging-data-events-with-cloudtrail.html) for certain services such as S3 objects and Lambda function invocations. Additional information about data event configuration can be found in the following links:

* [CloudTrail API DataResource documentation](https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_DataResource.html) (for basic event selector).
* [CloudTrail API AdvancedFieldSelector documentation](https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_AdvancedFieldSelector.html) (for advanced event selector).

#### Logging All Lambda Function Invocations By Using Basic Event Selectors

```terraform
resource "aws_cloudtrail" "example" {
  # ... other configuration ...

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type   = "AWS::Lambda::Function"
      values = ["arn:aws:lambda"]
    }
  }
}
```

#### Logging All S3 Object Events By Using Basic Event Selectors

```terraform
resource "aws_cloudtrail" "example" {
  # ... other configuration ...

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type   = "AWS::S3::Object"
      values = ["arn:aws:s3"]
    }
  }
}
```

#### Logging Individual S3 Bucket Events By Using Basic Event Selectors

```terraform
data "aws_s3_bucket" "important-bucket" {
  bucket = "important-bucket"
}

resource "aws_cloudtrail" "example" {
  # ... other configuration ...

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type = "AWS::S3::Object"

      # Make sure to append a trailing '/' to your ARN if you want
      # to monitor all objects in a bucket.
      values = ["${data.aws_s3_bucket.important-bucket.arn}/"]
    }
  }
}
```

#### Logging All S3 Object Events Except For Two S3 Buckets By Using Advanced Event Selectors

```terraform
data "aws_s3_bucket" "not-important-bucket-1" {
  bucket = "not-important-bucket-1"
}

data "aws_s3_bucket" "not-important-bucket-2" {
  bucket = "not-important-bucket-2"
}

resource "aws_cloudtrail" "example" {
  # ... other configuration ...

  advanced_event_selector {
    name = "Log all S3 objects events except for two S3 buckets"

    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field = "resources.ARN"

      not_equals = [
        "${data.aws_s3_bucket.not-important-bucket-1.arn}/",
        "${data.aws_s3_bucket.not-important-bucket-2.arn}/"
      ]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::S3::Object"]
    }
  }

  advanced_event_selector {
    name = "Log readOnly and writeOnly management events"

    field_selector {
      field  = "eventCategory"
      equals = ["Management"]
    }
  }
}
```

#### Logging Individual S3 Buckets And Specific Event Names By Using Advanced Event Selectors

```terraform
data "aws_s3_bucket" "important-bucket-1" {
  bucket = "important-bucket-1"
}

data "aws_s3_bucket" "important-bucket-2" {
  bucket = "important-bucket-2"
}

data "aws_s3_bucket" "important-bucket-3" {
  bucket = "important-bucket-3"
}

resource "aws_cloudtrail" "example" {
  # ... other configuration ...

  advanced_event_selector {
    name = "Log PutObject and DeleteObject events for two S3 buckets"

    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field = "eventName"

      equals = [
        "PutObject",
        "DeleteObject"
      ]
    }

    field_selector {
      field = "resources.ARN"

      #The trailing slash is intentional; do not exclude it.
      equals = [
        "${data.aws_s3_bucket.important-bucket-1.arn}/",
        "${data.aws_s3_bucket.important-bucket-2.arn}/"
      ]
    }

    field_selector {
      field  = "readOnly"
      equals = ["false"]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::S3::Object"]
    }
  }

  advanced_event_selector {
    name = "Log Delete* events for one S3 bucket"

    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field       = "eventName"
      starts_with = ["Delete"]
    }

    field_selector {
      field = "resources.ARN"

      #The trailing slash is intentional; do not exclude it.
      equals = [
        "${data.aws_s3_bucket.important-bucket-3.arn}/important-prefix"
      ]
    }

    field_selector {
      field  = "readOnly"
      equals = ["false"]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::S3::Object"]
    }
  }
}
```

#### Sending Events to CloudWatch Logs

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name = "Example"
}

resource "aws_cloudtrail" "example" {
  # ... other configuration ...

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.example.arn}:*" # CloudTrail requires the Log Stream wildcard
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the trail.
* `s3_bucket_name` - (Required) Name of the S3 bucket designated for publishing log files.

The following arguments are optional:

* `cloud_watch_logs_group_arn` - (Optional) Log group name using an ARN that represents the log group to which CloudTrail logs will be delivered. Note that CloudTrail requires the Log Stream wildcard.
* `cloud_watch_logs_role_arn` - (Optional) Role for the CloudWatch Logs endpoint to assume to write to a user’s log group.
* `enable_log_file_validation` - (Optional) Whether log file integrity validation is enabled. Defaults to `false`.
* `enable_logging` - (Optional) Enables logging for the trail. Defaults to `true`. Setting this to `false` will pause logging.
* `event_selector` - (Optional) Specifies an event selector for enabling data event logging. Fields documented below. Please note the [CloudTrail limits](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/WhatIsCloudTrail-Limits.html) when configuring these. Conflicts with `advanced_event_selector`.
* `advanced_event_selector` - (Optional) Specifies an advanced event selector for enabling data event logging. Fields documented below. Conflicts with `event_selector`.
* `include_global_service_events` - (Optional) Whether the trail is publishing events from global services such as IAM to the log files. Defaults to `true`.
* `insight_selector` - (Optional) Configuration block for identifying unusual operational activity. See details below.
* `is_multi_region_trail` - (Optional) Whether the trail is created in the current region or in all regions. Defaults to `false`.
* `is_organization_trail` - (Optional) Whether the trail is an AWS Organizations trail. Organization trails log events for the master account and all member accounts. Can only be created in the organization master account. Defaults to `false`.
* `kms_key_id` - (Optional) KMS key ARN to use to encrypt the logs delivered by CloudTrail.
* `s3_key_prefix` - (Optional) S3 key prefix that follows the name of the bucket you have designated for log file delivery.
* `sns_topic_name` - (Optional) Name of the Amazon SNS topic defined for notification of log file delivery.
* `tags` - (Optional) Map of tags to assign to the trail. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### event_selector

This configuration block supports the following attributes:

* `data_resource` - (Optional) Configuration block for data events. See details below.
* `exclude_management_event_sources` (Optional) -  A set of event sources to exclude. Valid values include: `kms.amazonaws.com` and `rdsdata.amazonaws.com`. `include_management_events` must be set to`true` to allow this.
* `include_management_events` - (Optional) Whether to include management events for your trail. Defaults to `true`.
* `read_write_type` - (Optional) Type of events to log. Valid values are `ReadOnly`, `WriteOnly`, `All`. Default value is `All`.

#### data_resource

This configuration block supports the following attributes:

* `type` - (Required) Resource type in which you want to log data events. You can specify only the following value: "AWS::S3::Object", "AWS::Lambda::Function" and "AWS::DynamoDB::Table".
* `values` - (Required) List of ARN strings or partial ARN strings to specify selectors for data audit events over data resources. ARN list is specific to single-valued `type`. For example, `arn:aws:s3:::<bucket name>/` for all objects in a bucket, `arn:aws:s3:::<bucket name>/key` for specific objects, `arn:aws:lambda` for all lambda events within an account, `arn:aws:lambda:<region>:<account number>:function:<function name>` for a specific Lambda function, `arn:aws:dynamodb` for all DDB events for all tables within an account, or `arn:aws:dynamodb:<region>:<account number>:table/<table name>` for a specific DynamoDB table.


### insight_selector

This configuration block supports the following attributes:

* `insight_type` - (Optional) Type of insights to log on a trail. Valid values are: `ApiCallRateInsight` and `ApiErrorRateInsight`.

### Advanced Event Selector Arguments
For **advanced_event_selector** the following attributes are supported.

* `name` (Optional) - Specifies the name of the advanced event selector.
* `field_selector` (Required) - Specifies the selector statements in an advanced event selector. Fields documented below.

#### Field Selector Arguments
For **field_selector** the following attributes are supported.

* `field` (Required) - Specifies a field in an event record on which to filter events to be logged. You can specify only the following values: `readOnly`, `eventSource`, `eventName`, `eventCategory`, `resources.type`, `resources.ARN`.
* `equals` (Optional) - A list of values that includes events that match the exact value of the event record field specified as the value of `field`. This is the only valid operator that you can use with the `readOnly`, `eventCategory`, and `resources.type` fields.
* `not_equals` (Optional) - A list of values that excludes events that match the exact value of the event record field specified as the value of `field`.
* `starts_with` (Optional) - A list of values that includes events that match the first few characters of the event record field specified as the value of `field`.
* `not_starts_with` (Optional) - A list of values that excludes events that match the first few characters of the event record field specified as the value of `field`.
* `ends_with` (Optional) - A list of values that includes events that match the last few characters of the event record field specified as the value of `field`.
* `not_ends_with` (Optional) - A list of values that excludes events that match the last few characters of the event record field specified as the value of `field`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the trail.
* `home_region` - Region in which the trail was created.
* `id` - Name of the trail.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Cloudtrails can be imported using the `name`, e.g.,

```
$ terraform import aws_cloudtrail.sample my-sample-trail
```
