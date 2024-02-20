# IVS (Interactive Video Service) Example

This example shows how to deploy an AWS IVS Chat room using Terraform only. The
example creates an AWS Lambda function for use as a message review handler as
well as an Amazon S3 bucket for chat logging. The handler will modify the
message by appending the text `- edited by Lambda` to the message.

To run, configure your AWS provider as described in https://www.terraform.io/docs/providers/aws/index.html

## Running the example

This example can be run by calling

```shell
terraform init
terraform apply
```

By default, resources are created in the `us-west-2` region. To override the
region, set the variable `aws_region` to a different value.
