---
subcategory: "Elemental MediaConvert"
layout: "aws"
page_title: "AWS: aws_media_convert_job_template"
description: |-
  Provides an AWS Elemental MediaConvert Job Template.
---

# Resource: aws_media_convert_job_template

Provides an AWS Elemental MediaConvert Job Template.

## Example Usage

```terraform
resource "aws_media_convert_job_template" "example" {
  name        = "tf-example-job-template"
  description = "Example MediaConvert job template"

  settings_json = jsonencode({
    outputGroups = [
      {
        name = "File Group"
        outputGroupSettings = {
          type = "FILE_GROUP_SETTINGS"
          fileGroupSettings = {
            destination = "s3://my-bucket/output/"
          }
        }
        outputs = [
          {
            containerSettings = {
              container = "MP4"
            }
            videoDescription = {
              codecSettings = {
                codec = "H_264"
                h264Settings = {
                  rateControlMode = "QVBR"
                }
              }
            }
          }
        ]
      }
    ]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required, Forces new resource) A unique name for the job template.
* `settings_json` - (Required) A JSON string representing the transcode settings for the job template. The JSON uses the same camelCase field names as the [MediaConvert API](https://docs.aws.amazon.com/mediaconvert/latest/apireference/jobtemplates.html).
* `acceleration_settings` - (Optional) Accelerated transcoding settings. See [`acceleration_settings`](#acceleration_settings) below.
* `category` - (Optional) A category for the job template.
* `description` - (Optional) A description of the job template.
* `hop_destinations` - (Optional) Optional list of hop destinations for queue hopping. See [`hop_destinations`](#hop_destinations) below.
* `priority` - (Optional) The relative priority for jobs created using this template. Valid values are `-50` to `50`. Defaults to `0`.
* `queue` - (Optional) The ARN of the default queue for jobs created from this template.
* `status_update_interval` - (Optional) How often MediaConvert sends STATUS_UPDATE events to Amazon CloudWatch Events. Valid values are `SECONDS_10`, `SECONDS_12`, `SECONDS_15`, `SECONDS_20`, `SECONDS_30`, `SECONDS_60`, `SECONDS_120`, `SECONDS_180`, `SECONDS_240`, `SECONDS_300`, `SECONDS_360`, `SECONDS_420`, `SECONDS_480`, `SECONDS_540`, `SECONDS_600`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### acceleration_settings

* `mode` - (Required) The acceleration mode. Valid values are `DISABLED`, `ENABLED`, or `PREFERRED`.

### hop_destinations

* `priority` - (Optional) The relative priority of the job in the destination queue.
* `queue` - (Optional) The ARN of the destination queue.
* `wait_minutes` - (Optional) The minimum wait time in minutes before the job can hop to the destination queue. Valid range is `1` to `4320`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the job template.
* `arn` - The ARN of the job template.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Media Convert Job Template using the template name. For example:

```terraform
import {
  to = aws_media_convert_job_template.example
  id = "tf-example-job-template"
}
```

Using `terraform import`, import Media Convert Job Template using the template name. For example:

```console
% terraform import aws_media_convert_job_template.example tf-example-job-template
```
