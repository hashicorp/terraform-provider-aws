---
subcategory: "Synthetics"
layout: "aws"
page_title: "AWS: aws_synthetics_canary"
description: |-
  Provides a Synthetics Canary resource
---

# Resource: aws_synthetics_canary

Provides a Synthetics Canary resource.

## Example Usage

```hcl
resource "aws_synthetics_canary" "some" {
  name                 = "some-canary"
  artifact_s3_location = "s3://some-bucket/"
  execution_role_arn   = "some-role"
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for this canary.
* `runtime_version` - (Required) Specifies the runtime version to use for the canary. Currently, the only valid values are `syn-nodejs-2.0`, `syn-nodejs-2.0-beta`, and `syn-1.0`. For more information about runtime versions, see [Canary Runtime Versions](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Synthetics_Canaries_Library.html).
* `artifact_s3_location` - (Required) The location on Amazon S3 where Synthetics stores artifacts from the test runs of this canary.
* `schedule` -  (Required) Information about how often the canary is to run and when these test runs are to stop. See [Schedule](#schedule) below.
* `handler` - (Required) The domain description.
* `execution_role_arn` - (Required) The ARN of the IAM role to be used to run the canary. see [AWS Docs](https://docs.aws.amazon.com/AmazonSynthetics/latest/APIReference/API_CreateCanary.html#API_CreateCanary_RequestSyntax) for permissions needs for IAM Role.
* `start_canary` - (Optional) Whether to run or stop the canary.
* `s3_bucket` - (Optional) If your canary script is located in S3, specify the full bucket name here. The bucket must already exist. Specify the full bucket name, including s3:// as the start of the bucket name. **Conflicts with `zip_file`**
* `s3_key` - (Optional) The S3 key of your script. **Conflicts with `zip_file`**
* `s3_version` - (Optional) The S3 version ID of your script. **Conflicts with `zip_file`**
* `zip_file` - (Optional)  If you input your canary script directly into the canary instead of referring to an S3 location, the value of this parameter is the .zip file that contains the script. It can be up to 5 MB. **Conflicts with `s3_bucket`, `s3_key`, and `s3_version`**
* `failure_retention_period` - (Optional) The number of days to retain data about failed runs of this canary. If you omit this field, the default of 31 days is used. The valid range is 1 to 455 days.
* `success_retention_period` - (Optional) The number of days to retain data about successful runs of this canary. If you omit this field, the default of 31 days is used. The valid range is 1 to 455 days.
* `run_config` - (Optional) Configuration for individual canary runs. See [Run Config](#run-config) below.
* `vpc_config` - (Optional) If this canary is to test an endpoint in a VPC, this structure contains information about the subnet and security groups of the VPC endpoint. For more information, see [Running a Canary in a VPC](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Synthetics_Canaries_VPC.html). See [VPC Config](#vpc-config) below.
* `tags` - (Optional) Key-value map of resource tags

### Schedule

* `expression` - (Required) A rate expression that defines how often the canary is to run. The syntax is rate(number unit). unit can be minute, minutes, or hour.
* `duration_in_seconds` - (Optional) Duration in seconds, for the canary to continue making regular runs according to the schedule in the Expression value.

### Run Config

* `timeout_in_seconds` - (Optional) How long the canary is allowed to run before it must stop. If you omit this field, the frequency of the canary is used as this value, up to a maximum of 14 minutes.
* `memory_in_mb` - (Optional) The maximum amount of memory available to the canary while it is running, in MB. The value you specify must be a multiple of 64.

### VPC Config

* `subnet_ids` - (Required) The IDs of the subnets where this canary is to run.
* `security_group_ids` - (Required) The IDs of the security groups for this canary.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name for this canary.
* `arn` - Amazon Resource Name (ARN) of the Canary.
* `source_location_arn` - The ARN of the Lambda layer where Synthetics stores the canary script code.
* `engine_arn` - The ARN of the Lambda function that is used as your canary's engine.
* `runtime_version` - Specifies the runtime version to use for the canary.
* `timeline` - A structure that contains information about when the canary was created, modified, and most recently run. see [Timeline](#timeline).

### VPC Config

* `vpc_id` - The ID of the VPC where this canary is to run.

### Timeline

* `created` - The date and time the canary was created.
* `last_modified` - The date and time the canary was most recently modified.
* `last_started` - The date and time that the canary's most recent run started.
* `last_stopped` - The date and time that the canary's most recent run ended.

## Import

Synthetics Canaries can be imported using the `name`, e.g.

```
$ terraform import aws_synthetics_canary.some some-canary
```

**Note about leftover implicit resources** - When a canary is created a set of resources are created implicitly,
 see [AWS Docs](https://docs.aws.amazon.com/AmazonSynthetics/latest/APIReference/API_DeleteCanary.html) for full list.
Terraform will not delete these resources automatically and require manual deletion or by using shell commands in terraform.
