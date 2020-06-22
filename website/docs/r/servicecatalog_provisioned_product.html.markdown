---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_provisioned_product"
description: |-
  Provides a resource for a Service Catalog Provisioned Product
  instance based on a Service Catalog Product definition.
---

# aws_servicecatalog_provisioned_product

Provides a resource for a Service Catalog Provisioned Product.

This represents an instance of a Service Catalog Product
which can be created, updated, and destroyed according to the product definition,
parameters, launch paths, and provisioning artifacts (e.g. versions).

## Example Usage

```hcl
resource "aws_servicecatalog_provisioned_product" "test" {
    provisioned_product_name = "my-new-product"
    product_id = "prod-abcdefghijklm"
    provisioning_artifact_id = "pa-0123456789abc"
}
```

## Argument Reference

The following arguments are supported:

* `notification_arns` - (Optional) SNS topic ARNs to which to publish stack-related events 
* `path_id` - (Optional) The path identifier of the product. This value is optional if the product has a default path, and required if the product has more than one path.
* `product_id` - (Required) The product identifier
* `provisioned_product_name` (Required) A user-friendly name for the provisioned product. This value must be unique for the AWS account and cannot be updated after the product is provisioned.
* `provisioning_artifact_id` (Required) The identifier of the provisioning artifact.
* `provisioning_parameters` (Optional) Parameters specified by the administrator that are required for provisioning the product. This object should be a map of the parameter key to the parameter value.
* `provisioning_preferences` - (Optional) An object that contains information about the provisioning preferences for a stack set. [Not yet supported.]
* `tags` - (Optional) Tags to apply to the provisioned product.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Service Catalog Provisioned Product.
* `created_time` - The UTC time stamp of the creation time.
* `outputs` - Information about the product created as the result of a request. For example, the output for a CloudFormation-backed product that creates an S3 bucket may include the S3 bucket URL. This is a map of output keys to their value (different to the AWS return type but allowing references to values in Terraform).
* `last_record_id` -  The ID of the last record read. Records provide a history of provsioned product activity.
* `last_record_status` -  The status reported by the last record read. The supported values are `CREATED | IN_PROGRESS | IN_PROGRESS_IN_ERROR | SUCCEEDED | FAILED`. This can sometimes be useful along with `last_record_type` to give more context to the provisioned product `status` and `status_message` attributes.
* `last_record_type` -  The type of the last record read. The supported values are `PROVISION_PRODUCT | UPDATE_PROVISIONED_PRODUCT | TERMINATE_PROVISIONED_PRODUCT`.
* `status` -  The status of the provisioned product. The supported values are `AVAILABLE | UNDER_CHANGE | TAINTED | ERROR | PLAN_IN_PROGRESS`.
* `status_message` -  The current status message of the provisioned product.
* `updated_time` - The time when the provisioned product (and record) was last updated.

## Import

Service Catalog Provisioned Products can be imported using the Service Catalog Provisioned Product ID, e.g.

```
$ terraform import aws_servicecatalog_provisioned_product.test p-12344321
```
