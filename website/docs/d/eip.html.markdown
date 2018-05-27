---
layout: "aws"
page_title: "AWS: aws_eip"
sidebar_current: "docs-aws-datasource-eip"
description: |-
    Provides details about a specific Elastic IP
---

# Data Source: aws_eip

`aws_eip` provides details about a specific Elastic IP.

This resource can prove useful when a module accepts an allocation ID or
public IP as an input variable and needs to determine the other.

## Example Usage

The following example shows how one might accept a public IP as a variable
and use this data source to obtain the allocation ID.

```hcl
variable "instance_id" {}
variable "public_ip" {}

data "aws_eip" "proxy_ip" {
  public_ip = "${var.public_ip}"
}

resource "aws_eip_association" "proxy_eip" {
  instance_id   = "${var.instance_id}"
  allocation_id = "${data.aws_eip.proxy_ip.id}"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Elastic IPs in the current region. The given filters must match exactly one
Elastic IP whose data will be exported as attributes.

* `id` - (Optional) The allocation id of the specific EIP to retrieve.

* `public_ip` - (Optional) The public IP of the specific EIP to retrieve.

* `eip_tags` - (Optional) A mapping of tags, each pair of which must
exactly match a pair on the desired EIP to retrieve.

* `filter` - (Optional) One or more name/value pairs to use as filters. There are
several valid keys, for a full reference, check out
[describe-addresses in the AWS CLI reference][1].

## Attributes Reference

All of the argument attributes are also exported as result attributes. This
data source will complete the data by populating any fields that are not
included in the configuration with the data for the selected Elastic IP.

[1]: https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-addresses.html
