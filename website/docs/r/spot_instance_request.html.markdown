---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_spot_instance_request"
description: |-
  Provides a Spot Instance Request resource.
---

# Resource: aws_spot_instance_request

Provides an EC2 Spot Instance Request resource. This allows instances to be
requested on the spot market.

By default Terraform creates Spot Instance Requests with a `persistent` type,
which means that for the duration of their lifetime, AWS will launch an
instance with the configured details if and when the spot market will accept
the requested price.

On destruction, Terraform will make an attempt to terminate the associated Spot
Instance if there is one present.

Spot Instances requests with a `one-time` type will close the spot request
when the instance is terminated either by the request being below the current spot
price availability or by a user.

~> **NOTE:** Because their behavior depends on the live status of the spot
market, Spot Instance Requests have a unique lifecycle that makes them behave
differently than other Terraform resources. Most importantly: there is __no
guarantee__ that a Spot Instance exists to fulfill the request at any given
point in time. See the [AWS Spot Instance
documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-instances.html)
for more information.

~> **NOTE [AWS strongly discourages](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-best-practices.html#which-spot-request-method-to-use) the use of the legacy APIs called by this resource.
We recommend using the [EC2 Instance](instance.html) resource with `instance_market_options` instead.

## Example Usage

```terraform
# Request a spot instance at $0.03
resource "aws_spot_instance_request" "cheap_worker" {
  ami           = "ami-1234"
  spot_price    = "0.03"
  instance_type = "c4.xlarge"

  tags = {
    Name = "CheapWorker"
  }
}
```

## Argument Reference

Spot Instance Requests support all the same arguments as
[`aws_instance`](instance.html), with the addition of:

* `spot_price` - (Optional; Default: On-demand price) The maximum price to request on the spot market.
* `wait_for_fulfillment` - (Optional; Default: false) If set, Terraform will
  wait for the Spot Request to be fulfilled, and will throw an error if the
  timeout of 10m is reached.
* `spot_type` - (Optional; Default: `persistent`) If set to `one-time`, after
  the instance is terminated, the spot request will be closed.
* `launch_group` - (Optional) A launch group is a group of spot instances that launch together and terminate together.
  If left empty instances are launched and terminated individually.
* `block_duration_minutes` - (Optional) The required duration for the Spot instances, in minutes. This value must be a multiple of 60 (60, 120, 180, 240, 300, or 360).
  The duration period starts as soon as your Spot instance receives its instance ID. At the end of the duration period, Amazon EC2 marks the Spot instance for termination and provides a Spot instance termination notice, which gives the instance a two-minute warning before it terminates.
  Note that you can't specify an Availability Zone group or a launch group if you specify a duration.
* `instance_interruption_behavior` - (Optional) Indicates Spot instance behavior when it is interrupted. Valid values are `terminate`, `stop`, or `hibernate`. Default value is `terminate`.
* `valid_until` - (Optional) The end date and time of the request, in UTC [RFC3339](https://tools.ietf.org/html/rfc3339#section-5.8) format(for example, YYYY-MM-DDTHH:MM:SSZ). At this point, no new Spot instance requests are placed or enabled to fulfill the request. The default end date is 7 days from the current date.
* `valid_from` - (Optional) The start date and time of the request, in UTC [RFC3339](https://tools.ietf.org/html/rfc3339#section-5.8) format(for example, YYYY-MM-DDTHH:MM:SSZ). The default is to start fulfilling the request immediately.
* `tags` - (Optional) A map of tags to assign to the Spot Instance Request. These tags are not automatically applied to the launched Instance. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Spot Instance Request ID.

These attributes are exported, but they are expected to change over time and so
should only be used for informational purposes, not for resource dependencies:

* `spot_bid_status` - The current [bid
  status](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-bid-status.html)
  of the Spot Instance Request.
* `spot_request_state` The current [request
  state](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-requests.html#creating-spot-request-status)
  of the Spot Instance Request.
* `spot_instance_id` - The Instance ID (if any) that is currently fulfilling
  the Spot Instance request.
* `public_dns` - The public DNS name assigned to the instance. For EC2-VPC, this
  is only available if you've enabled DNS hostnames for your VPC
* `public_ip` - The public IP address assigned to the instance, if applicable.
* `private_dns` - The private DNS name assigned to the instance. Can only be
  used inside the Amazon EC2, and only available if you've enabled DNS hostnames
  for your VPC
* `private_ip` - The private IP address assigned to the instance
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `read` - (Default `15m`)
* `delete` - (Default `20m`)
