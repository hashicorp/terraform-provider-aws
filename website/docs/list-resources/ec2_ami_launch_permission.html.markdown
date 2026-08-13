---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_ami_launch_permission"
description: |-
  Lists EC2 (Elastic Compute Cloud) AMI Launch Permission resources.
---

# List Resource: aws_ec2_ami_launch_permission

Lists EC2 (Elastic Compute Cloud) AMI Launch Permission resources.

This list resource returns all launch permissions granted on AMIs owned by the caller's account.

## Example Usage

### Basic Usage

```terraform
list "aws_ec2_ami_launch_permission" "example" {
  provider = aws
}
```

### Including Resource Data

```terraform
list "aws_ec2_ami_launch_permission" "example" {
  provider = aws

  include_resource = true
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
