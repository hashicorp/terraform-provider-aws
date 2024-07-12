---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_metadata_defaults"
description: |-
  Manages regional EC2 instance metadata default settings.
---

# Resource: aws_ec2_instance_metadata_defaults

Manages regional EC2 instance metadata default settings.
More information can be found in the [Configure instance metadata options for new instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-IMDS-new-instances.html) user guide.

## Example Usage

```terraform
resource "aws_ec2_instance_metadata_defaults" "enforce-imdsv2" {
  http_tokens                 = "required"
  http_put_response_hop_limit = 1
}
```

## Argument Reference

This resource supports the following arguments.

* `http_endpoint` - (Optional) Whether the metadata service is available. Can be `"enabled"`, `"disabled"`, or `"no-preference"`. Default: `"no-preference"`.
* `http_tokens` - (Optional) Whether the metadata service requires session tokens, also referred to as _Instance Metadata Service Version 2 (IMDSv2)_. Can be `"optional"`, `"required"`, or `"no-preference"`. Default: `"no-preference"`.
* `http_put_response_hop_limit` - (Optional) The desired HTTP PUT response hop limit for instance metadata requests. The larger the number, the further instance metadata requests can travel. Can be an integer from `1` to `64`, or `-1` to indicate no preference. Default: `-1`.
* `instance_metadata_tags` - (Optional) Enables or disables access to instance tags from the instance metadata service. Can be `"enabled"`, `"disabled"`, or `"no-preference"`. Default: `"no-preference"`.

## Attribute Reference

This data source exports no additional attributes.

## Import

You cannot import this resource.
