---
layout: aws
page_title: 'AWS: aws_guardduty_ipset'
sidebar_current: docs-aws-resource-guardduty-ipset
description: Provides a resource to manage a GuardDuty IPSet
---

# aws_guardduty_ipset

Provides a resource to manage a GuardDuty detector.

## Example Usage

```hcl
resource "aws_guardduty_detector" "master" {
  enable = true
}

resource "aws_s3_bucket" "bucket" {
  acl = "private"
}

resource "aws_s3_bucket_object" "MyIPSet" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = "${aws_s3_bucket.bucket.id}"
  key     = "MyIPSet"
}

resource "aws_guardduty_ipset" "MyIPSet" {
  activate    = true
  detector_id = "${aws_guardduty_detector.master.id}"
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_bucket_object.MyIPSet.bucket}/${aws_s3_bucket_object.MyIPSet.key}"
  name        = "MyIPSet"
}
```

## Argument Reference

The following arguments are supported:

* `activate` - (Required) Specifies whether GuardDuty is to start using the uploaded IPSet.
* `detector_id` - (Required) The detector ID of the GuardDuty.
* `format` - (Required) The format of the file that contains the IPSet. Valid values: `TXT` | `STIX` | `OTX_CSV` | `ALIEN_VAULT` | `PROOF_POINT` | `FIRE_EYE`
* `location` - (Required) The URI of the file that contains the IPSet.
* `name` - (Required) The friendly name to identify the IPSet.

## Attributes Reference

The following additional attributes are exported:

* `id` - The ID of the GuardDuty IPSet.
