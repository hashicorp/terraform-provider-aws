---
subcategory: "Resource Explorer"
layout: "aws"
page_title: "AWS: aws_resourceexplorer2_search"
description: |-
  Terraform data source for managing an AWS Resource Explorer Search.
---
# Data Source: aws_resourceexplorer2_search

Terraform data source for managing an AWS Resource Explorer Search.

## Example Usage

### Basic Usage

```terraform
data "aws_resourceexplorer2_search" "example" {
  query_string = "region:us-west-2"
  view_arn     = aws_resourceexplorer2_index.test.arn
}
```

## Argument Reference

The following arguments are required:

* `query_string` - (Required) String that includes keywords and filters that specify the resources that you want to include in the results. For the complete syntax supported by the QueryString parameter, see Search query syntax reference for [Resource Explorer](https://docs.aws.amazon.com/resource-explorer/latest/userguide/using-search-query-syntax.html). The search is completely case insensitive. You can specify an empty string to return all results up to the limit of 1,000 total results. The operation can return only the first 1,000 results. If the resource you want is not included, then use a different value for QueryString to refine the results.

The following arguments are optional:

* `view_arn` - (Optional) Specifies the Amazon resource name (ARN) of the view to use for the query. If you don't specify a value for this parameter, then the operation automatically uses the default view for the AWS Region in which you called this operation. If the Region either doesn't have a default view or if you don't have permission to use the default view, then the operation fails with a `401 Unauthorized` exception.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `resource_count` - Number of resources that match the query. See [`resource_count`](#resource_count-attribute-reference) below.
* `resources` - List of structures that describe the resources that match the query. See [`resources`](#resources-attribute-reference) below.
* `id` - Query String.

### `resource_count` Attribute Reference

* `complete` - Indicates whether the TotalResources value represents an exhaustive count of search results. If True, it indicates that the search was exhaustive. Every resource that matches the query was counted. If False, then the search reached the limit of 1,000 matching results, and stopped counting.
* `total_resources` - Number of resources that match the search query. This value can't exceed 1,000. If there are more than 1,000 resources that match the query, then only 1,000 are counted and the Complete field is set to false. We recommend that you refine your query to return a smaller number of results.

### `resources` Attribute Reference

* `arn` - Amazon resource name of resource.
* `last_reported_at` - Date and time that Resource Explorer last queried this resource and updated the index with the latest information about the resource.
* `owning_account_id` - Amazon Web Services account that owns the resource.
* `properties` - Structure with additional type-specific details about the resource.  See [`properties`](#properties-attribute-reference) below.
* `region` - Amazon Web Services Region in which the resource was created and exists.
* `resource_type` - Type of the resource.
* `service` - Amazon Web Service that owns the resource and is responsible for creating and updating it.

### `properties` Attribute Reference

* `data` - Details about this property. The content of this field is a JSON object that varies based on the resource type.
* `last_reported_at` - The date and time that the information about this resource property was last updated.
* `name` - Name of this property of the resource.
