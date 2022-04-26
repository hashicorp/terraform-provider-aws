# Direct Connect Gateway Cross-Account VGW Association

This example demonstrates how to create a VPN Gateway in one AWS account, create a Direct Connect Gateway in a second AWS account, and associate the VPN Gateway with the Direct Connect Gateway via the `aws_dx_gateway_association_proposal` and `aws_dx_gateway_association` resources.

See [more in the Direct Connect Gateway documentation](https://docs.aws.amazon.com/directconnect/latest/UserGuide/direct-connect-gateways.html).

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
