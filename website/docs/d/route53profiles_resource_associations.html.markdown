---
subcategory: "Route 53 Profiles"
layout: "aws"
page_title: "AWS: aws_route53profiles_resource_associations"
description: |-
  Provides details about the resources associated with a Route 53 Profile.
---

# Data Source: aws_route53profiles_resource_associations

Provides details about the resources associated with a Route 53 Profile, such as private hosted zones, Resolver rules, DNS Firewall rule groups, interface VPC endpoints and Resolver query logging configurations.

If nothing matches, `resource_associations` is an empty list.

## Example Usage

### Basic Usage

```terraform
data "aws_route53profiles_resource_associations" "example" {
  profile_id = aws_route53profiles_profile.example.id
}
```

### DNS Firewall Rule Groups

```terraform
data "aws_route53profiles_resource_associations" "example" {
  profile_id    = aws_route53profiles_profile.example.id
  resource_type = "FIREWALL_RULE_GROUP"
}

locals {
  firewall_rule_group_arns = [for a in data.aws_route53profiles_resource_associations.example.resource_associations : a.resource_arn]
}
```

## Argument Reference

The following arguments are required:

* `profile_id` - (Required) ID of the Profile.

The following arguments are optional:

* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_type` - (Optional) Type of resource to return, for example `HOSTED_ZONE` or `FIREWALL_RULE_GROUP`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `resource_associations` - List of resource associations of the Profile. Each element contains:
    * `id` - ID of the resource association.
    * `name` - Name of the resource association.
    * `owner_id` - AWS account ID of the resource association owner.
    * `profile_id` - ID of the Profile.
    * `resource_arn` - ARN of the associated resource.
    * `resource_properties` - JSON-encoded properties of the resource association, such as the priority of a DNS Firewall rule group.
    * `resource_type` - Type of the associated resource.
    * `status` - Status of the resource association.
    * `status_message` - Status message of the resource association.
