# aws-go-sdk-base

[https://github.com/hashicorp/aws-sdk-go-base](https://github.com/hashicorp/aws-sdk-go-base)

This is a base library used by the [AWS Provider](https://github.com/hashicorp/terraform-provider-aws), [AWSCC Provider](https://github.com/hashicorp/terraform-provider-awscc) and the [Terraform S3 Backend](https://github.com/hashicorp/terraform/tree/main/internal/backend/remote-state/s3) to allow them to handle authentication and other non-service level AWS interactions consistently.

Typically this changes infrequently and changes are normally performed by HashiCorp maintainers. It should not be necessary to change this library for the majority of provider contributions.
