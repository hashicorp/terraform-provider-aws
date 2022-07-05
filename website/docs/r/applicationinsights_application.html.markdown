---
subcategory: "CloudWatch Application Insights"
layout: "aws"
page_title: "AWS: aws_applicationinsights_application"
description: |-
  Provides a CloudWatch Application Insights Application resource
---

# Resource: aws_applicationinsights_application

Provides a ApplicationInsights Application resource.

## Example Usage

```terraform
resource "aws_applicationinsights_application" "example" {
  resource_group_name = aws_resourcegroups_group.example.name
}

resource "aws_resourcegroups_group" "example" {
  name = "example"

  resource_query {
    query = <<JSON
	{
		"ResourceTypeFilters": [
		  "AWS::EC2::Instance"
		],
		"TagFilters": [
		  {
			"Key": "Stage",
			"Values": [
			  "Test"
			]
		  }
		]
	  }
JSON
  }
}
```

## Argument Reference

The following arguments are required:

* `resource_group_name` - (Required) The name of the resource group.

The following arguments are optional:

* `auto_config_enabled` - (Optional)  Indicates whether Application Insights automatically configures unmonitored resources in the resource group.
* `auto_create` - (Optional) Configures all of the resources in the resource group by applying the recommended configurations.
* `cwe_monitor_enabled` - (Optional)  Indicates whether Application Insights can listen to CloudWatch events for the application resources, such as instance terminated, failed deployment, and others.
* `grouping_type` - (Optional) Application Insights can create applications based on a resource group or on an account. To create an account-based application using all of the resources in the account, set this parameter to `ACCOUNT_BASED`.
* `ops_center_enabled` - (Optional) When set to `true`, creates opsItems for any problems detected on an application.
* `ops_item_sns_topic_arn` - (Optional) The SNS topic provided to Application Insights that is associated to the created opsItem. Allows you to receive notifications for updates to the opsItem.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the Application.
* `id` - The name of the resource group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).


## Import

ApplicationInsights Applications can be imported using the `resource_group_name`, e.g.,

```
$ terraform import aws_applicationinsights_application.some some-application
```
