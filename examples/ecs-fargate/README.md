# ECS with Fargate

This example shows how to launch an ECS service fronted with Application Load Balancer running on Fargate.

__Note:__ Fargate may not be available in your region. You can verify this by reading https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services

To run, configure your AWS provider as described in https://www.terraform.io/docs/providers/aws/index.html

## Get up and running

Planning phase

```
terraform plan
```

Apply phase

```
terraform apply
```

Once the stack is created, wait for a few minutes and test the stack by launching a browser with the ALB url.

## Destroy :boom:

```
terraform destroy
```
