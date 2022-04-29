---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_vpc_connector"
description: |-
  Manages an App Runner VPC Connector.
---

# Resource: aws_apprunner_vpc_connector

Manages an App Runner VPC Connector.

## Example Usage

```terraform
resource "aws_apprunner_vpc_connector" "connector" {
  vpc_connector_name = "name"
  subnets            = ["subnet1", "subnet2"]
  security_groups    = ["sg1", "sg2"]
}
```

## Argument Reference

The following arguments supported:

* `vpc_connector_name` - (Required) A name for the VPC connector.
* `subnets` (Required) A list of IDs of subnets that App Runner should use when it associates your service with a custom Amazon VPC. Specify IDs of subnets of a single Amazon VPC. App Runner determines the Amazon VPC from the subnets you specify.
* `security_groups` - A list of IDs of security groups that App Runner should use for access to AWS resources under the specified subnets. If not specified, App Runner uses the default security group of the Amazon VPC. The default security group allows all outbound traffic.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `vpc_connector_arn` - The Amazon Resource Name (ARN) of VPC connector.
* `status` - The current state of the VPC connector. If the status of a connector revision is INACTIVE, it was deleted and can't be used. Inactive connector revisions are permanently removed some time after they are deleted.
* `vpc_connector_revision` - The revision of VPC connector. It's unique among all the active connectors ("Status": "ACTIVE") that share the same Name.

## Import

App Runner vpc connector can be imported by using the `vpc_connector_name`, e.g.,

```
$ terraform import aws_apprunner_vpc_connector.vpc_connector_name 
```
