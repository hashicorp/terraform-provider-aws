---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_anycast_ip_list"
description: |-
  Terraform resource for managing a CloudFront Anycast IP List.
---

# Resource: aws_cloudfront_anycast_ip_list

Terraform resource for managing a CloudFront Anycast IP List.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudfront_anycast_ip_list" "example" {
  name     = "example-list"
  ip_count = 21
}
```

### Bring Your Own IP (BYOIP) via IPAM

Allocate the Anycast static IPs from address space you have brought into Amazon VPC IPAM instead of the AWS-owned pool.

```terraform
resource "aws_cloudfront_anycast_ip_list" "example" {
  name     = "example-list"
  ip_count = 3

  ipam_cidr_config {
    cidr          = "203.0.113.0/24"
    ipam_pool_arn = aws_vpc_ipam_pool.example.arn
    anycast_ip    = "203.0.113.10"
  }
}
```

## Argument Reference

The following arguments are required:

* `ip_count` - (Required, Forces new resource) The number of static IP addresses that are allocated to the Anycast IP list. Valid values: `3`, `21`.
* `name` - (Required, Forces new resource) Name of the Anycast IP list.

The following arguments are optional:

* `ipam_cidr_config` - (Optional, Forces new resource) Configuration block for one or more IPAM CIDRs used to allocate the Anycast static IPs from your own address space (BYOIP). [See below](#ipam_cidr_config).
* `tags` - (Optional) Key-value tags for the place index. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### ipam_cidr_config

* `cidr` - (Required, Forces new resource) CIDR that specifies the IP address range for this IPAM configuration.
* `ipam_pool_arn` - (Required, Forces new resource) Amazon Resource Name (ARN) of the IPAM pool that the CIDR block is assigned to.
* `anycast_ip` - (Optional, Forces new resource) Specific Anycast IP address to allocate from the IPAM pool for this CIDR configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `anycast_ips` - The static IP addresses that are allocated to the Anycast IP list.
* `arn` - The Anycast IP list ARN.
* `etag` - The current version of the Anycast IP list.
* `id` - The Anycast IP list ID.
* `ipam_config` - IPAM configuration for the Anycast IP list. [See below](#ipam_config).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### ipam_config

* `quantity` - Number of IPAM CIDR configurations.
* `ipam_cidr_configs` - List of IPAM CIDR configurations. Each entry contains:
    * `cidr` - CIDR for this IPAM configuration.
    * `ipam_pool_arn` - ARN of the IPAM pool that the CIDR block is assigned to.
    * `anycast_ip` - Anycast IP address allocated from the IPAM pool for this CIDR configuration.
    * `status` - Current status of the IPAM CIDR configuration (for example, `provisioning`, `provisioned`, `advertising`, `advertised`).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Anycast IP List using the `id`. For example:

```terraform
import {
  to = aws_cloudfront_anycast_ip_list.example
  id = "abcd-1234"
}
```

Using `terraform import`, import CloudFront Anycast IP List using the `id`. For example:

```console
% terraform import aws_cloudfront_anycast_ip_list.example abcd-1234 
```
