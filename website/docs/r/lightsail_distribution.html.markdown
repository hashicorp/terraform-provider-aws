---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_distribution"
description: |-
  Manages a Lightsail content delivery network (CDN) distribution.
---

# Resource: aws_lightsail_distribution

Manages a Lightsail content delivery network (CDN) distribution. Use this resource to cache content at edge locations and reduce latency for users accessing your content.

## Example Usage

### Basic Usage

```terraform
resource "aws_lightsail_bucket" "example" {
  name      = "example-bucket"
  bundle_id = "small_1_0"
}

resource "aws_lightsail_distribution" "example" {
  name      = "example-distribution"
  bundle_id = "small_1_0"
  origin {
    name        = aws_lightsail_bucket.example.name
    region_name = aws_lightsail_bucket.example.region
  }
  default_cache_behavior {
    behavior = "cache"
  }
  cache_behavior_settings {
    allowed_http_methods = "GET,HEAD,OPTIONS,PUT,PATCH,POST,DELETE"
    cached_http_methods  = "GET,HEAD"
    default_ttl          = 86400
    maximum_ttl          = 31536000
    minimum_ttl          = 0
    forwarded_cookies {
      option = "none"
    }
    forwarded_headers {
      option = "default"
    }
    forwarded_query_strings {
      option = false
    }
  }
}
```

### Instance Origin

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_static_ip_attachment" "example" {
  static_ip_name = aws_lightsail_static_ip.example.name
  instance_name  = aws_lightsail_instance.example.name
}

resource "aws_lightsail_static_ip" "example" {
  name = "example-static-ip"
}

resource "aws_lightsail_instance" "example" {
  name              = "example-instance"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "micro_1_0"
}

resource "aws_lightsail_distribution" "example" {
  name       = "example-distribution"
  depends_on = [aws_lightsail_static_ip_attachment.example]
  bundle_id  = "small_1_0"
  origin {
    name        = aws_lightsail_instance.example.name
    region_name = data.aws_availability_zones.available.id
  }
  default_cache_behavior {
    behavior = "cache"
  }
}
```

### Load Balancer Origin

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_lb" "example" {
  name              = "example-load-balancer"
  health_check_path = "/"
  instance_port     = "80"
  tags = {
    foo = "bar"
  }
}

resource "aws_lightsail_instance" "example" {
  name              = "example-instance"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}

resource "aws_lightsail_lb_attachment" "example" {
  lb_name       = aws_lightsail_lb.example.name
  instance_name = aws_lightsail_instance.example.name
}

resource "aws_lightsail_distribution" "example" {
  name       = "example-distribution"
  depends_on = [aws_lightsail_lb_attachment.example]
  bundle_id  = "small_1_0"
  origin {
    name        = aws_lightsail_lb.example.name
    region_name = data.aws_availability_zones.available.id
  }
  default_cache_behavior {
    behavior = "cache"
  }
}
```

## Argument Reference

The following arguments are required:

* `bundle_id` - (Required) Bundle ID to use for the distribution.
* `default_cache_behavior` - (Required) Default cache behavior of the distribution. [See below](#default_cache_behavior).
* `name` - (Required) Name of the distribution.
* `origin` - (Required) Origin resource of the distribution, such as a Lightsail instance, bucket, or load balancer. [See below](#origin).

The following arguments are optional:

* `cache_behavior` - (Optional) Per-path cache behavior of the distribution. [See below](#cache_behavior).
* `cache_behavior_settings` - (Optional) Cache behavior settings of the distribution. [See below](#cache_behavior_settings).
* `certificate_name` - (Optional) Name of the SSL/TLS certificate attached to the distribution.
* `ip_address_type` - (Optional) IP address type of the distribution. Valid values: `dualstack`, `ipv4`. Default: `dualstack`.
* `is_enabled` - (Optional) Whether the distribution is enabled. Default: `true`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags for the Lightsail Distribution. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### cache_behavior

* `behavior` - (Required) Cache behavior for the specified path. Valid values: `cache`, `dont-cache`.
* `path` - (Required) Path to a directory or file to cache, or not cache. Use an asterisk symbol to specify wildcard directories (`path/to/assets/*`), and file types (`*.html`, `*.jpg`, `*.js`). Directories and file paths are case-sensitive.

### cache_behavior_settings

* `allowed_http_methods` - (Optional) HTTP methods that are processed and forwarded to the distribution's origin.
* `cached_http_methods` - (Optional) HTTP method responses that are cached by your distribution.
* `default_ttl` - (Optional) Default amount of time that objects stay in the distribution's cache before the distribution forwards another request to the origin to determine whether the content has been updated.
* `forwarded_cookies` - (Optional) Cookies that are forwarded to the origin. Your content is cached based on the cookies that are forwarded. [See below](#forwarded_cookies).
* `forwarded_headers` - (Optional) Headers that are forwarded to the origin. Your content is cached based on the headers that are forwarded. [See below](#forwarded_headers).
* `forwarded_query_strings` - (Optional) Query strings that are forwarded to the origin. Your content is cached based on the query strings that are forwarded. [See below](#forwarded_query_strings).
* `maximum_ttl` - (Optional) Maximum amount of time that objects stay in the distribution's cache before the distribution forwards another request to the origin to determine whether the object has been updated.
* `minimum_ttl` - (Optional) Minimum amount of time that objects stay in the distribution's cache before the distribution forwards another request to the origin to determine whether the object has been updated.

#### forwarded_cookies

* `cookies_allow_list` - (Optional) Specific cookies to forward to your distribution's origin.
* `option` - (Optional) Which cookies to forward to the distribution's origin for a cache behavior. Valid values: `all`, `none`, `allow-list`.

#### forwarded_headers

* `headers_allow_list` - (Optional) Specific headers to forward to your distribution's origin.
* `option` - (Optional) Headers that you want your distribution to forward to your origin and base caching on. Valid values: `default`, `allow-list`, `all`.

#### forwarded_query_strings

* `option` - (Optional) Whether the distribution forwards and caches based on query strings.
* `query_strings_allowed_list` - (Optional) Specific query strings that the distribution forwards to the origin.

### default_cache_behavior

* `behavior` - (Required) Cache behavior of the distribution. Valid values: `cache`, `dont-cache`.

### origin

* `name` - (Required) Name of the origin resource. Your origin can be an instance with an attached static IP, a bucket, or a load balancer that has at least one instance attached to it.
* `protocol_policy` - (Optional) Protocol that your Amazon Lightsail distribution uses when establishing a connection with your origin to pull content.
* `region_name` - (Required) AWS Region name of the origin resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `alternative_domain_names` - Alternate domain names of the distribution.
* `arn` - ARN of the distribution.
* `created_at` - Timestamp when the distribution was created.
* `domain_name` - Domain name of the distribution.
* `location` - Location of the distribution, such as the AWS Region and Availability Zone. [See below](#location).
* `origin_public_dns` - Public DNS of the origin.
* `origin[0].resource_type` - Resource type of the origin resource (e.g., Instance).
* `resource_type` - Lightsail resource type (e.g., Distribution).
* `status` - Status of the distribution.
* `support_code` - Support code. Include this code in your email to support when you have questions about your Lightsail distribution. This code enables our support team to look up your Lightsail information more easily.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### location

* `availability_zone` - Availability Zone. Follows the format us-east-2a (case-sensitive).
* `region_name` - AWS Region name.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lightsail Distribution using the `name`. For example:

```terraform
import {
  to = aws_lightsail_distribution.example
  id = "example-distribution"
}
```

Using `terraform import`, import Lightsail Distribution using the `name`. For example:

```console
% terraform import aws_lightsail_distribution.example example-distribution
```
