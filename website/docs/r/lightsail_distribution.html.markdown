---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_distribution"
description: |-
  Terraform resource for managing an AWS Lightsail Distribution.
---

# Resource: aws_lightsail_distribution

Terraform resource for managing an AWS Lightsail Distribution.

## Example Usage

### Basic Usage

Below is a basic example with a bucket as an origin.

```terraform
resource "aws_lightsail_bucket" "test" {
  name      = "test-bucket"
  bundle_id = "small_1_0"
}
resource "aws_lightsail_distribution" "test" {
  name      = "test-distribution"
  bundle_id = "small_1_0"
  origin {
    name        = aws_lightsail_bucket.test.name
    region_name = aws_lightsail_bucket.test.region
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

### instance origin example

Below is an example of an instance as the origin.

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_static_ip_attachment" "test" {
  static_ip_name = aws_lightsail_static_ip.test.name
  instance_name  = aws_lightsail_instance.test.name
}

resource "aws_lightsail_static_ip" "test" {
  name = "test-static-ip"
}

resource "aws_lightsail_instance" "test" {
  name              = "test-instance"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "micro_1_0"
}

resource "aws_lightsail_distribution" "test" {
  name       = "test-distribution"
  depends_on = [aws_lightsail_static_ip_attachment.test]
  bundle_id  = "small_1_0"
  origin {
    name        = aws_lightsail_instance.test.name
    region_name = data.aws_availability_zones.available.id
  }
  default_cache_behavior {
    behavior = "cache"
  }
}
```

### lb origin example

Below is an example with a load balancer as an origin

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_lb" "test" {
  name              = "test-load-balancer"
  health_check_path = "/"
  instance_port     = "80"
  tags = {
    foo = "bar"
  }
}

resource "aws_lightsail_instance" "test" {
  name              = "test-instance"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}

resource "aws_lightsail_lb_attachment" "test" {
  lb_name       = aws_lightsail_lb.test.name
  instance_name = aws_lightsail_instance.test.name
}

resource "aws_lightsail_distribution" "test" {
  name       = "test-distribution"
  depends_on = [aws_lightsail_lb_attachment.test]
  bundle_id  = "small_1_0"
  origin {
    name        = aws_lightsail_lb.test.name
    region_name = data.aws_availability_zones.available.id
  }
  default_cache_behavior {
    behavior = "cache"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the distribution.
* `bundle_id` - (Required) Bundle ID to use for the distribution.
* `default_cache_behavior` - (Required) Object that describes the default cache behavior of the distribution. [Detailed below](#default_cache_behavior)
* `origin` - (Required) Object that describes the origin resource of the distribution, such as a Lightsail instance, bucket, or load balancer. [Detailed below](#origin)
* `cache_behavior_settings` - (Required) An object that describes the cache behavior settings of the distribution. [Detailed below](#cache_behavior_settings)

The following arguments are optional:

* `cache_behavior` - (Optional) A set of configuration blocks that describe the per-path cache behavior of the distribution. [Detailed below](#cache_behavior)
* `certificate_name` - (Optional) The name of the SSL/TLS certificate attached to the distribution, if any.
* `ip_address_type` - (Optional) The IP address type of the distribution. Default: `dualstack`.
* `is_enabled` - (Optional) Indicates whether the distribution is enabled. Default: `true`.
* `tags` - (Optional) Map of tags for the Lightsail Distribution. To create a key-only tag, use an empty string as the value. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### default_cache_behavior

* `behavior` - (Required) The cache behavior of the distribution. Valid values: `cache` and `dont-cache`.

### origin

* `name` - (Required) The name of the origin resource. Your origin can be an instance with an attached static IP, a bucket, or a load balancer that has at least one instance attached to it.
* `protocol_policy` - (Optional) The protocol that your Amazon Lightsail distribution uses when establishing a connection with your origin to pull content.
* `region_name` - (Required) The AWS Region name of the origin resource.
* `resource_type` - (Computed) The resource type of the origin resource (e.g., Instance).

### cache_behavior_settings

* `allowed_http_methods` - (Optional) The HTTP methods that are processed and forwarded to the distribution's origin.
* `cached_http_methods` - (Optional) The HTTP method responses that are cached by your distribution.
* `default_ttl` - (Optional) The default amount of time that objects stay in the distribution's cache before the distribution forwards another request to the origin to determine whether the content has been updated.
* `forwarded_cookies` - (Required) An object that describes the cookies that are forwarded to the origin. Your content is cached based on the cookies that are forwarded. [Detailed below](#forwarded_cookies)
* `forwarded_headers` - (Required) An object that describes the headers that are forwarded to the origin. Your content is cached based on the headers that are forwarded. [Detailed below](#forwarded_headers)
* `forwarded_query_strings` - (Required) An object that describes the query strings that are forwarded to the origin. Your content is cached based on the query strings that are forwarded. [Detailed below](#forwarded_query_strings)
* `maximum_ttl` - (Optional) The maximum amount of time that objects stay in the distribution's cache before the distribution forwards another request to the origin to determine whether the object has been updated.
* `minimum_ttl` - (Optional) The minimum amount of time that objects stay in the distribution's cache before the distribution forwards another request to the origin to determine whether the object has been updated.

#### forwarded_cookies

* `cookies_allow_list` - (Required) The specific cookies to forward to your distribution's origin.
* `option` - (Optional) Specifies which cookies to forward to the distribution's origin for a cache behavior: all, none, or allow-list to forward only the cookies specified in the cookiesAllowList parameter.

#### forwarded_headers

* `headers_allow_list` - (Required) The specific headers to forward to your distribution's origin.
* `option` - (Optional) The headers that you want your distribution to forward to your origin and base caching on.

#### forwarded_query_strings

* `option` - (Optional) Indicates whether the distribution forwards and caches based on query strings.
* `query_strings_allowed_list` - (Required) The specific query strings that the distribution forwards to the origin.

### cache_behavior

* `behavior` - (Required) The cache behavior for the specified path.
* `path` - (Required) The path to a directory or file to cached, or not cache. Use an asterisk symbol to specify wildcard directories (path/to/assets/\*), and file types (\*.html, \*jpg, \*js). Directories and file paths are case-sensitive.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `alternative_domain_names` - The alternate domain names of the distribution.
* `arn` - The Amazon Resource Name (ARN) of the distribution.
* `created_at` - The timestamp when the distribution was created.
* `domain_name` - The domain name of the distribution.
* `location` - An object that describes the location of the distribution, such as the AWS Region and Availability Zone. [Detailed below](#location)
* `origin_public_dns` - The public DNS of the origin.
* `resource_type` - The Lightsail resource type (e.g., Distribution).
* `status` - The status of the distribution.
* `support_code` - The support code. Include this code in your email to support when you have questions about your Lightsail distribution. This code enables our support team to look up your Lightsail information more easily.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### location

* `availability_zone` - The Availability Zone. Follows the format us-east-2a (case-sensitive).
* `region_name` - The AWS Region name.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lightsail Distribution using the `id`. For example:

```terraform
import {
  to = aws_lightsail_distribution.example
  id = "rft-8012925589"
}
```

Using `terraform import`, import Lightsail Distribution using the `id`. For example:

```console
% terraform import aws_lightsail_distribution.example rft-8012925589
```
