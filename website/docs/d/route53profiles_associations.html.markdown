---
subcategory: "Route 53 Profiles"
layout: "aws"
page_title: "AWS: aws_route53profiles_associations"
description: |-
  Provides details about Route 53 Profile associations with VPCs.
---

# Data Source: aws_route53profiles_associations

Provides details about the associations between Route 53 Profiles and VPCs in the current Region.

The arguments filter the results:

* Neither `profile_id` nor `resource_id`: every Profile-VPC association in the Region.
* `profile_id` only: the associations of that Profile, one for each VPC.
* `resource_id` only: the associations of that VPC, one for each Profile.
* Both: the association between that Profile and that VPC.

If nothing matches, `associations` is an empty list.

## Example Usage

### Basic Usage

```terraform
data "aws_route53profiles_associations" "example" {
  resource_id = aws_vpc.example.id
}

locals {
  used_priorities = [for a in data.aws_route53profiles_associations.example.associations : a.priority]
}
```

### Associations of a Profile

```terraform
data "aws_route53profiles_associations" "example" {
  profile_id = aws_route53profiles_profile.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `profile_id` - (Optional) ID of the Profile. Returns only the VPC associations of this Profile.
* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_id` - (Optional) ID of the VPC. Returns only the Profile associations of this VPC.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `associations` - List of Profile-VPC associations. Each element contains:
    * `id` - ID of the Profile association.
    * `name` - Name of the Profile association.
    * `owner_id` - AWS account ID of the Profile association owner.
    * `priority` - Priority of the Profile association. VPC Resolver evaluates a VPC's Profiles from the lowest value.
    * `profile_id` - ID of the Profile.
    * `resource_id` - ID of the VPC the Profile is associated with.
    * `status` - Status of the Profile association.
    * `status_message` - Status message of the Profile association.
