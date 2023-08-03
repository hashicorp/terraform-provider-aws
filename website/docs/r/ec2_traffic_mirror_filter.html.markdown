---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_traffic_mirror_filter"
description: |-
  Provides an Traffic mirror filter
---

# Resource: aws_ec2_traffic_mirror_filter

Provides an Traffic mirror filter.  
Read [limits and considerations](https://docs.aws.amazon.com/vpc/latest/mirroring/traffic-mirroring-considerations.html) for traffic mirroring

## Example Usage

To create a basic traffic mirror filter

```terraform
resource "aws_ec2_traffic_mirror_filter" "foo" {
  description      = "traffic mirror filter - terraform example"
  network_services = ["amazon-dns"]
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional, Forces new resource) A description of the filter.
* `network_services` - (Optional) List of amazon network services that should be mirrored. Valid values: `amazon-dns`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the traffic mirror filter.
* `id` - The name of the filter.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import traffic mirror filter using the `id`. For example:

```terraform
import {
  to = aws_ec2_traffic_mirror_filter.foo
  id = "tmf-0fbb93ddf38198f64"
}
```

Using `terraform import`, import traffic mirror filter using the `id`. For example:

```console
% terraform import aws_ec2_traffic_mirror_filter.foo tmf-0fbb93ddf38198f64
```
