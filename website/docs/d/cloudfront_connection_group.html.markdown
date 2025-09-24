---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_connection_group"
description: |-
  Provides a CloudFront connection group data source.
---

# Data Source: aws_cloudfront_connection_group

Use this data source to retrieve information about a CloudFront connection group.

## Example Usage

```terraform
data "aws_cloudfront_connection_group" "test" {
  id = "EDFDVBD632BHDS5"
}
```

## Argument Reference

Exactly one of the following arguments must be specified for this data source:

* `id` (Optional) - Identifier for the connection group. For example: `EDFDVBD632BHDS5`. 
* `routing_endpoint` (Optional) - Routing endpoint for the connection group. For example: `d111111abcdef8.cloudfront.net`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `anycast_ip_list_id` - ID of the anycast IP list associated with this connection group, if any.

* `arn` - ARN (Amazon Resource Name) for the connection group.

* `status` - Current status of the connection group. `Deployed` if the
    connection group's information is fully propagated throughout the Amazon
    CloudFront system.

* `enabled` - Whether the connection group is enabled.

* `last_modified_time` - Date and time the connection group was last modified.

* `is_default` - Whether the connection group is the default connection group for the distribution tenants.

* `etag` - Current version of the connection group's information. For example:
    `E2QWRUHAPOMQZL`.

* `name` - name of the connection group.
