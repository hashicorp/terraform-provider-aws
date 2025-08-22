# IVS (Interactive Video Service) Example

This example shows how to deploy an AWS IVS channel using Terraform only. The
example creates an S3 bucket for recording.

To run, configure your AWS provider as described in https://www.terraform.io/docs/providers/aws/index.html

## Running the example

Run `terraform apply` to see it work.

By default, resources are created in the `us-west-2` region. To override the
region, set the variable `aws_region` to a different value.
