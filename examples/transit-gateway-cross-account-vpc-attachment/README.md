# EC2 Transit Gateway Cross-Account VPC Attachment

This example demonstrates how to create a Transit Gateway in one AWS account, share it with a second AWS account, and attach a VPC in the second account to the Transit Gateway.

See [more in the Transit Gateway documentation](https://docs.aws.amazon.com/vpc/latest/tgw/tgw-transit-gateways.html).

## Running this example

Either `cp terraform.template.tfvars terraform.tfvars` and modify that new file accordingly or provide variables via CLI:

```
terraform apply \
	-var="aws_first_access_key=AAAAAAAAAAAAAAAAAAA" \
	-var="aws_first_secret_key=SuperSecretKeyForAccount1" \
	-var="aws_second_access_key=BBBBBBBBBBBBBBBBBBB" \
	-var="aws_second_secret_key=SuperSecretKeyForAccount2" \
	-var="aws_region=us-east-1"
```

## Prerequisite 
- In the AWS console of the owner of the TGW, enabling AWS Resource Access Manager by going into `Services`, `Organization`, `Settings` and `Enable Access` to AWS Resource Access Manager. 
- In the AWS console of the owner of the TGW, enabling Resources Access Manager by going into `Services`, `Resource Access Manager`, `Settings` and tickle the `Enable sharing within your AWS Organization`. 
