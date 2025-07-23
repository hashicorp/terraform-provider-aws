---
subcategory: "EBS (EC2)"
layout: "aws"
page_title: "AWS: aws_ebs_snapshot_ids"
description: |-
  Provides a list of EBS snapshot IDs.
---

# Data Source: aws_ebs_snapshot_ids

Use this data source to get a list of EBS Snapshot IDs matching the specified
criteria.

## Example Usage

```terraform
data "aws_ebs_snapshot_ids" "ebs_volumes" {
  owners = ["self"]

  filter {
    name   = "volume-size"
    values = ["40"]
  }

  filter {
    name   = "tag:Name"
    values = ["Example"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `owners` - (Optional) Returns the snapshots owned by the specified owner id. Multiple owners can be specified.
* `restorable_by_user_ids` - (Optional) One or more AWS accounts IDs that can create volumes from the snapshot.
* `filter` - (Optional) One or more name/value pairs to filter off of. There are several valid keys, for a full reference, check out [describe-volumes in the AWS CLI reference][1].

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `ids` - Set of EBS snapshot IDs, sorted by creation time in descending order.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

[1]: http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-snapshots.html
