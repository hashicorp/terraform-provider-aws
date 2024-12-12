# Amazon OpenSearch Serverless Example

This example creates an OpenSearch Serverless collection with an encryption, network, and data access policy along with an OpenSearch Serverless-managed VPC endpoint and its required VPC resources.

## Setup

Configure your AWS provider as described in https://www.terraform.io/docs/providers/aws/index.html

From inside of this directory:

```bash
export AWS_ACCESS_KEY_ID=<this is a secret>
export AWS_SECRET_ACCESS_KEY=<this is a secret>
terraform init
```

## Run the example

From inside of this directory:

```bash
terraform apply
```

## Clean up

From inside of this directory:

```bash
terraform destroy
```
