---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_entity_path"
description: |-
  Get the entity path for an entity.
---

# Data Source: aws_organizations_entity_path

Get the [entity path](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_last-accessed-view-data-orgs.html#access_policies_last-accessed-viewing-orgs-entity-path) for an entity. An entity's path is the text representation of the structure of that AWS Organizations entity.

## Example Usage

```terraform
data "aws_organizations_entity_path" "example" {
  entity_id = "ou-ghi0-awsccccc"
}
```

## Argument Reference

This data source supports the following arguments:

* `entity_id` - (Required) Entity ID. Must be an organizational unit (OU) or AWS account ID.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `entity_path` - Entity path.
