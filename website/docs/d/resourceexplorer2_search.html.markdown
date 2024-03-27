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

* `query_string` - (Required) String that includes keywords and filters that specify the resources that you want to include in the results. For the complete syntax supported by the QueryString parameter, see Search query syntax reference for Resource Explorer (https://docs.aws.amazon.com/resource-explorer/latest/userguide/using-search-query-syntax.html). The search is completely case insensitive. You can specify an empty string to return all results up to the limit of 1,000 total results. The operation can return only the first 1,000 results. If the resource you want is not included, then use a different value for QueryString to refine the results.

The following arguments are optional:

* `max_results` - (Optional) Maximum number of results that you want included on each page of the response. If you do not include this parameter, it defaults to a value appropriate to the operation.

* `next_token` - (Optional) Parameter for receiving additional results if you receive a `NextToken` response in a previous request. A NextToken response indicates that more output is available.

* `view_arn` - (Optional) Specifies the Amazon resource name (ARN) of the view to use for the query. If you don't specify a value for this parameter, then the operation automatically uses the default view for the AWS Region in which you called this operation. If the Region either doesn't have a default view or if you don't have permission to use the default view, then the operation fails with a `401 Unauthorized` exception.
