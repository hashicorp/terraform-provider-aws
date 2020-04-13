---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_spot_price_history"
description: |-
  Show price development.
---

# Resource: aws_spot_price_history

Show the development of spot instance prices. This can be used in conjunction with a launch configuration.
 
See the [AWS Spot Instance
documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-instances.html)
for more information.


## Example Usage

```hcl
# Request spot prices
resource "aws_spot_price_history" "prices" {
  start_time = "2020-04-13T01:10:30Z"
  filter {
    name   = "instance_type"
    values = ["m4.large"]
  }
}

# use the latest price for the launch configuration
resource "aws_launch_configuration" "default" {
  name          = "default"
  image_id      = "${data.aws_ami.ubuntu.id}"
  instance_type = "m4.large"
  spot_price    = data.aws_spot_price_history.prices.latest.spot_price * 1.2
}

```

## Argument Reference

* `start_time` - (Optional) The date and time, up to the past 90 days, from which to start retrieving the price history data, in UTC format (for example, YYYY-MM-DDTHH:MM:SSZ).
* `end_time` - (Optional) The date and time, up to the current date, from which to stop retrieving the price history data, in UTC format (for example, YYYY-MM-DDTHH:MM:SSZ).
* `filter` - (Optional) One or more name/value pairs to filter off of. There are several valid keys, 
for a full reference, check out [describe-spot-price-history in the AWS CLI reference][1].

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - a random ID.
* `latest` is the latest recorded spot price data. See [Spot Price Data](#spot-price-data)
* `previous` is a list of older prices. See [Spot Price Data](#spot-price-data) for he format of each element

## Spot Price Data

* `availability_zone` - The zome where this instance type can be started 
* `instance_type` -  The instance type
* `product_description` - The product (Linux/UNIX | SUSE Linux | Windows |)
* `spot_price` - The price
* `timestamp` - The time from which the price has become available 


[1]: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html
