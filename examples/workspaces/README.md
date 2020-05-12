# AWS WorkSpaces Example

This example demonstrates how to create a WorkSpace and WorkSpace directory using AWS WorkSpaces.

## Note

The AWS WorkSpaces service requires an IAM role named `workspaces_DefaultRole`. This example creates this role and attaches policies to it. This will cause an error if the role already exists in the AWS account.

The IAM resources are defined in the Terraform source file [iam.tf](./iam.tf). If the role exists, remove the IAM resource file. The resources `aws_workspaces_directory.example` and `aws_workspaces_workspace.example` have dependencies on the IAM resources. The `depends_on` blocks can be removed from both.

## Running this example

This example can be run by calling

```shell
terraform init
terraform apply
```

By default, resources are created in the `us-west-2` region. To override the region, set the variable `aws_region` to a different value.
