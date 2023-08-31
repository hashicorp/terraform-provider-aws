---
subcategory: "GuardDuty"
layout: aws
page_title: 'AWS: aws_guardduty_threatintelset'
description: Provides a resource to manage a GuardDuty ThreatIntelSet
---

# Resource: aws_guardduty_threatintelset

Provides a resource to manage a GuardDuty ThreatIntelSet.

~> **Note:** Currently in GuardDuty, users from member accounts cannot upload and further manage ThreatIntelSets. ThreatIntelSets that are uploaded by the primary account are imposed on GuardDuty functionality in its member accounts. See the [GuardDuty API Documentation](https://docs.aws.amazon.com/guardduty/latest/ug/create-threat-intel-set.html)

## Example Usage

```terraform
resource "aws_guardduty_detector" "primary" {
  enable = true
}

resource "aws_s3_bucket" "bucket" {
  # ... other configuration ...
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

resource "aws_s3_object" "MyThreatIntelSet" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = aws_s3_bucket.bucket.id
  key     = "MyThreatIntelSet"
}

resource "aws_guardduty_threatintelset" "MyThreatIntelSet" {
  activate    = true
  detector_id = aws_guardduty_detector.primary.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_object.MyThreatIntelSet.bucket}/${aws_s3_object.MyThreatIntelSet.key}"
  name        = "MyThreatIntelSet"
}
```

## Argument Reference

This resource supports the following arguments:

* `activate` - (Required) Specifies whether GuardDuty is to start using the uploaded ThreatIntelSet.
* `detector_id` - (Required) The detector ID of the GuardDuty.
* `format` - (Required) The format of the file that contains the ThreatIntelSet. Valid values: `TXT` | `STIX` | `OTX_CSV` | `ALIEN_VAULT` | `PROOF_POINT` | `FIRE_EYE`
* `location` - (Required) The URI of the file that contains the ThreatIntelSet.
* `name` - (Required) The friendly name to identify the ThreatIntelSet.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the GuardDuty ThreatIntelSet.
* `id` - The ID of the GuardDuty ThreatIntelSet and the detector ID. Format: `<DetectorID>:<ThreatIntelSetID>`
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import GuardDuty ThreatIntelSet using the primary GuardDuty detector ID and ThreatIntelSetID. For example:

```terraform
import {
  to = aws_guardduty_threatintelset.MyThreatIntelSet
  id = "00b00fd5aecc0ab60a708659477e9617:123456789012"
}
```

Using `terraform import`, import GuardDuty ThreatIntelSet using the primary GuardDuty detector ID and ThreatIntelSetID. For example:

```console
% terraform import aws_guardduty_threatintelset.MyThreatIntelSet 00b00fd5aecc0ab60a708659477e9617:123456789012
```
