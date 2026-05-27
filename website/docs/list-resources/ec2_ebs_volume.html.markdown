---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_ebs_volume"
description: |-
  Lists EC2 (Elastic Compute Cloud) EBS Volume resources.
---

# List Resource: aws_ec2_ebs_volume

Lists EC2 (Elastic Compute Cloud) EBS Volume resources.

## Example Usage

### Basic Usage

```terraform
list "aws_ec2_ebs_volume" "example" {
  provider = aws
}
```

### Filter Usage

This example returns EBS Volumes in a specific Availability Zone.

```terraform
list "aws_ec2_ebs_volume" "example" {
  provider = aws

  config {
    filter {
      name   = "availability-zone"
      values = ["us-west-2a"]
    }
  }
}
```

This example returns EBS Volumes with a specific tag value.

```terraform
list "aws_ec2_ebs_volume" "example" {
  provider = aws

  config {
    filter {
      name   = "tag:Name"
      values = ["example-volume"]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filter` - (Optional) One or more filters to apply to the search.
  If multiple `filter` blocks are provided, they all must be true.
  For a full reference of filter names, see [describe-volumes in the AWS CLI reference][describe-volumes].
  See [`filter` Block](#filter-block) below.
* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query.
  Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `volume_ids` - (Optional) List of EBS Volume IDs to query.

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter.
  For a full reference of filter names, see [describe-volumes in the AWS CLI reference][describe-volumes].
* `values` - (Required) One or more values to match.

[describe-volumes]: http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-volumes.html
