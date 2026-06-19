---
subcategory: "Connect Campaigns V2"
layout: "aws"
page_title: "AWS: aws_connectcampaignsv2_campaign"
description: |-
  Provides an Amazon Connect Campaigns V2 Campaign resource.
---

# Resource: aws_connectcampaignsv2_campaign

Provides an Amazon Connect Campaigns V2 Campaign resource.

~> This initial implementation exposes the campaign core arguments plus native Terraform blocks for `entry_limits_config`, `schedule`, and `source`. Additional complex campaign configuration shapes will be added as native nested schema in a future enhancement.

## Example Usage

```terraform
resource "aws_connectcampaignsv2_campaign" "example" {
  connect_instance_id = aws_connect_instance.example.id
  name                = "example-campaign"
  type                = "AGENTLESS"

  entry_limits_config {
    max_entry_count    = 1
    min_entry_interval = "PT24H"
  }

  source {
    customer_profiles_segment_arn = "arn:aws:profile:us-east-1:123456789012:domains/domain/segments/segment"
  }

  tags = {
    Name = "example-campaign"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.
* `connect_campaign_flow_arn` - (Optional) Amazon Resource Name (ARN) of the Amazon Connect campaign flow. Removing this value after creation is not supported by this resource.
* `connect_instance_id` - (Required, Forces new resource) Amazon Connect instance identifier.
* `entry_limits_config` - (Optional) Campaign entry limits config. Removing this block after creation deletes the campaign entry limits configuration. See [entry_limits_config](#entry_limits_config) below.
* `name` - (Required) Name of the campaign.
* `schedule` - (Optional) Campaign schedule. Removing this block after creation is not supported by this resource. See [schedule](#schedule) below.
* `source` - (Optional) Campaign source. Use either `customer_profiles_segment_arn` or `event_trigger`. Removing this block after creation is not supported by this resource. See [source](#source) below.
* `tags` - (Optional) Tags to apply to the Campaign. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional, Forces new resource) Type of campaign externally exposed in APIs.

### entry_limits_config

The `entry_limits_config` block supports the following arguments:

* `max_entry_count` - (Required) Maximum number of times a participant can enter the campaign. Use `0` for unlimited entries.
* `min_entry_interval` - (Required) Minimum time interval that must pass before a participant can enter the campaign again, in ISO 8601 duration format.

### schedule

The `schedule` block supports the following arguments:

* `end_time` - (Required) Campaign end time in RFC3339 format.
* `refresh_frequency` - (Optional) Campaign refresh frequency in ISO 8601 duration format.
* `start_time` - (Required) Campaign start time in RFC3339 format.

### source

The `source` block supports the following arguments:

* `customer_profiles_segment_arn` - (Optional) Customer Profiles segment ARN source for the campaign. Exactly one of this argument or `event_trigger` must be configured.

The `event_trigger` block supports the following arguments:

* `customer_profiles_domain_arn` - (Optional) Customer Profiles domain ARN for an event-triggered campaign source.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Campaign.
* `id` - Identifier of the Campaign.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.

## Import

In Terraform v1.5.0 and later, use an `import` block to import Amazon Connect Campaigns V2 Campaigns using the campaign ID. For example:

```terraform
import {
  to = aws_connectcampaignsv2_campaign.example
  id = "12345678-1234-1234-1234-123456789012"
}
```
