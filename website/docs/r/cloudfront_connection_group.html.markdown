---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_connection_group"
description: |-
  Provides a CloudFront connection group resource.
---

# Resource: aws_cloudfront_connection_group

Creates an Amazon CloudFront Connection Group.

For information about CloudFront Connection Groups, see the [AWS CloudFormation Documentation][1].

## Example Usage

```terraform
resource "aws_cloudfront_connection_group" "example" {
  name         = "example"
  enabled      = true
  ipv6_enabled = true
}
```

## Argument Reference

This resource supports the following arguments:

* `anycast_ip_list_id` (Optional) - ID of the associated Anycast IP List. `ipv6_enabled` must not be set to `true` in order to set this argument
* `enabled` (Optional) - Whether the connection group is enabled. Default is `true`.
* `ipv6_enabled` (Optional) - Whether IPv6 is enabled for the connection group. Default is `false`.
* `name` (Required) - Name of the connection group.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the connection group.
* `arn` - ARN of the connection group.
* `status` - Current status of the connection group.
* `is_default` - Whether the connection group is the default connection group for the distribution tenants.
* `routing_endpoint` - The routing endpoint (also known as the DNS name) that is assigned to the connection group, such as d111111abcdef8.cloudfront.net.
* `last_modified_time` - Date and time when the connection group was last modified.
* `etag` - Current version of the connection group.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Connection Groups using the `id`. For example:

```terraform
import {
  to = aws_cloudfront_connection_group.example
  id = "CGROUPT123EXAMPLE"
}
```

Using `terraform import`, import CloudFront Connection Groups using the `id`. For example:

```console
% terraform import aws_cloudfront_copnection_group.example CGROUP123EXAMPLE
```

[1]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-cloudfront-connectiongroup.html
