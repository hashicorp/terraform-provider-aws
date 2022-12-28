---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_instance"
description: |-
  Provides an Lightsail Instance
---

# Resource: aws_lightsail_instance

Provides a Lightsail Instance. Amazon Lightsail is a service to provide easy virtual private servers
with custom software already setup. See [What is Amazon Lightsail?](https://lightsail.aws.amazon.com/ls/docs/getting-started/article/what-is-amazon-lightsail)
for more information.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage

### Basic Usage

```terraform
# Create a new GitLab Lightsail Instance
resource "aws_lightsail_instance" "gitlab_test" {
  name              = "custom_gitlab"
  availability_zone = "us-east-1b"
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  key_pair_name     = "some_key_name"
  tags = {
    foo = "bar"
  }
}
```

### Enable Auto Snapshots

```terraform
resource "aws_lightsail_instance" "test" {
  name              = "custom_instance"
  availability_zone = "us-east-1b"
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
  add_on {
    type          = "AutoSnapshot"
    snapshot_time = "06:00"
    status        = "Enabled"
  }
  tags = {
    foo = "bar"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Lightsail Instance. Names be unique within each AWS Region in your Lightsail account.
* `availability_zone` - (Required) The Availability Zone in which to create your
instance (see list below)
* `blueprint_id` - (Required) The ID for a virtual private server image. A list of available blueprint IDs can be obtained using the AWS CLI command: `aws lightsail get-blueprints`
* `bundle_id` - (Required) The bundle of specification information (see list below)
* `key_pair_name` - (Optional) The name of your key pair. Created in the
Lightsail console (cannot use `aws_key_pair` at this time)
* `user_data` - (Optional) launch script to configure server with additional user data
* `ip_address_type` - (Optional) The IP address type of the Lightsail Instance. Valid Values: `dualstack` | `ipv4`.
* `add_on` - (Optional) The add on configuration for the instance. [Detailed below](#add_on).
* `tags` - (Optional) A map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `add_on`

Defines the add on configuration for the instance. The `add_on` configuration block supports the following arguments:

* `type` - (Required) The add-on type. There is currently only one valid type `AutoSnapshot`.
* `snapshot_time` - (Required) The daily time when an automatic snapshot will be created. Must be in HH:00 format, and in an hourly increment and specified in Coordinated Universal Time (UTC). The snapshot will be automatically created between the time specified and up to 45 minutes after.
* `status` - (Required) The status of the add on. Valid Values: `Enabled`, `Disabled`.

## Availability Zones
Lightsail currently supports the following Availability Zones (e.g., `us-east-1a`):

- `ap-northeast-1{a,c,d}`
- `ap-northeast-2{a,c}`
- `ap-south-1{a,b}`
- `ap-southeast-1{a,b,c}`
- `ap-southeast-2{a,b,c}`
- `ca-central-1{a,b}`
- `eu-central-1{a,b,c}`
- `eu-west-1{a,b,c}`
- `eu-west-2{a,b,c}`
- `eu-west-3{a,b,c}`
- `us-east-1{a,b,c,d,e,f}`
- `us-east-2{a,b,c}`
- `us-west-2{a,b,c}`

## Bundles

Lightsail currently supports the following Bundle IDs (e.g., an instance in `ap-northeast-1` would use `small_2_0`):

### Prefix

A Bundle ID starts with one of the below size prefixes:

- `nano_`
- `micro_`
- `small_`
- `medium_`
- `large_`
- `xlarge_`
- `2xlarge_`

### Suffix

A Bundle ID ends with one of the following suffixes depending on Availability Zone:

- ap-northeast-1: `2_0`
- ap-northeast-2: `2_0`
- ap-south-1: `2_1`
- ap-southeast-1: `2_0`
- ap-southeast-2: `2_2`
- ca-central-1: `2_0`
- eu-central-1: `2_0`
- eu-west-1: `2_0`
- eu-west-2: `2_0`
- eu-west-3: `2_0`
- us-east-1: `2_0`
- us-east-2: `2_0`
- us-west-2: `2_0`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the Lightsail instance (matches `arn`).
* `arn` - The ARN of the Lightsail instance (matches `id`).
* `created_at` - The timestamp when the instance was created.
* `cpu_count` - The number of vCPUs the instance has.
* `ram_size` - The amount of RAM in GB on the instance (e.g., 1.0).
* `ipv6_address` - (**Deprecated**) The first IPv6 address of the Lightsail instance. Use `ipv6_addresses` attribute instead.
* `ipv6_addresses` - List of IPv6 addresses for the Lightsail instance.
* `private_ip_address` - The private IP address of the instance.
* `public_ip_address` - The public IP address of the instance.
* `is_static_ip` - A Boolean value indicating whether this instance has a static IP assigned to it.
* `username` - The user name for connecting to the instance (e.g., ec2-user).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Lightsail Instances can be imported using their name, e.g.,

```
$ terraform import aws_lightsail_instance.gitlab_test 'custom_gitlab'
```
