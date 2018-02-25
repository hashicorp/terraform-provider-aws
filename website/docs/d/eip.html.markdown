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
and use this data source to obtain the allocation ID when using an VPC EIP.

### ip or id
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

### filter

```hcl
variable "instance_id" {}

resource "aws_eip" "proxy_ip" {
  vpc = true

  tags {
      Name = "proxy"
    }
}

data "aws_eip" "by_filter" {
  filter {
    name   = "tag:Name"
    values = ["${aws_eip.test.tags.Name}"]
  }
}

resource "aws_eip_association" "proxy_eip" {
  instance_id   = "${var.instance_id}"
  allocation_id = "${data.aws_eip.proxy_ip.id}"
}

```

~> **NOTE:** if using `data "aws_eip"` on a none pre exsiting EIP, ensure you reference the tag of the EIP by interpolation as shown in the example.

Classic EIP's do not have an allocation_id, only use `public_ip` in the `data "aws_eip"` block.

## Argument Reference

The arguments of this data source act as filters for querying the available
Elastic IPs in the current region. The given filters must match exactly one
Elastic IP whose data will be exported as attributes.

* `id` - (Optional) The allocation id of the specific VPC EIP to retrieve. If a classic EIP is required, do NOT set `id`, only set `public_ip`

* `public_ip` - (Optional) The public IP of the specific EIP to retrieve.

* `filter` - (Optional) One or more name/value pairs to use as filters. There are
several valid keys, for a full reference, check out
[describe-addresses in the AWS CLI reference][1].

~> **NOTE:** `filter` cannot be used when `id` or `public_ip` is being used.

## Attributes Reference

All of the argument attributes are also exported as result attributes. This
data source will complete the data by populating any fields that are not
included in the configuration with the data for the selected Elastic IP.

