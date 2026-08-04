---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_volume_attachment"
description: |-
  Lists EBS Volume Attachment resources.
---

# List Resource: aws_volume_attachment

Lists EBS Volume Attachment resources.

## Example Usage

```terraform
list "aws_volume_attachment" "example" {
  config {
    instance_id = aws_instance.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `instance_id` - (Required) ID of the EC2 Instance to list volume attachments for.
* `region` - (Optional) Region to query. Defaults to provider region.
