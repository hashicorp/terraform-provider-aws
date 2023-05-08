# EC2 Transit Gateway intra-region Peering

This example demonstrates how to create two Transit Gateways in one AWS account and the same region, attach a VPC each, and then create a Peering Attachment between the two Transit Gateways.

See [more in the Transit Gateway intra-region Peering documentation](https://aws.amazon.com/it/blogs/networking-and-content-delivery/aws-transit-gateway-now-supports-intra-region-peering/).

## Running this example

Either `cp terraform.template.tfvars terraform.tfvars` and modify that new file accordingly or provide variables via CLI:

```terrform
terraform apply \
	-var="aws_profile=aws-account" \
	-var="aws_region=us-east-1" 
```

## Prerequisites

- This example requires one AWS accounts within the same AWS Organizations Organization
- Ensure Resource Access Manager is enabled in your organization. For more information, see the [Resource Access Manager User Guide](https://docs.aws.amazon.com/ram/latest/userguide/getting-started-sharing.html).
