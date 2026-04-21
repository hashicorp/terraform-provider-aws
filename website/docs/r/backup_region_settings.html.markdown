---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_region_settings"
description: |-
  Provides an AWS Backup Region Settings resource.
---

# Resource: aws_backup_region_settings

Provides an AWS Backup Region Settings resource.

## Example Usage

```terraform
resource "aws_backup_region_settings" "test" {
  resource_type_opt_in_preference = {
    "Aurora"                 = true
    "CloudFormation"         = true
    "DocumentDB"             = true
    "DSQL"                   = true
    "DynamoDB"               = true
    "EBS"                    = true
    "EC2"                    = true
    "EFS"                    = true
    "FSx"                    = true
    "Neptune"                = true
    "Redshift"               = true
    "Redshift Serverless"    = false
    "RDS"                    = false
    "S3"                     = false
    "SAP HANA on Amazon EC2" = false
    "Storage Gateway"        = false
    "VirtualMachine"         = false
  }

  resource_type_management_preference = {
    "CloudFormation" = true
    "DSQL"           = true
    "DynamoDB"       = false
    "EFS"            = false
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_type_opt_in_preference` - (Required) A map of service names to their opt-in preferences for the Region. See [AWS Documentation on which services support backup](https://docs.aws.amazon.com/aws-backup/latest/devguide/backup-feature-availability.html).
* `resource_type_management_preference` - (Optional) A map of service names to their full management preferences for the Region. For more information, see the AWS Documentation on [what full management is](https://docs.aws.amazon.com/aws-backup/latest/devguide/whatisbackup.html#full-management) and [which services support full management](https://docs.aws.amazon.com/aws-backup/latest/devguide/backup-feature-availability.html#features-by-resource).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The AWS region.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup Region Settings using the `region`. For example:

```terraform
import {
  to = aws_backup_region_settings.test
  id = "us-west-2"
}
```

Using `terraform import`, import Backup Region Settings using the `region`. For example:

```console
% terraform import aws_backup_region_settings.test us-west-2
```
