---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_resource"
description: |-
  Get information on a API Gateway Resource
---

# Data Source: aws_api_gateway_resource

Use this data source to get the id of a Resource in API Gateway.
To fetch the Resource, you must provide the REST API id as well as the full path.  

## Example Usage

```terraform
data "aws_api_gateway_rest_api" "my_rest_api" {
  name = "my-rest-api"
}

data "aws_api_gateway_resource" "my_resource" {
  rest_api_id = data.aws_api_gateway_rest_api.my_rest_api.id
  path        = "/endpoint/path"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rest_api_id` - (Required) REST API id that owns the resource. If no REST API is found, an error will be returned.
* `path` - (Required) Full path of the resource.  If no path is found, an error will be returned.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Set to the ID of the found Resource.
* `parent_id` - Set to the ID of the parent Resource.
* `path_part` - Set to the path relative to the parent Resource.
