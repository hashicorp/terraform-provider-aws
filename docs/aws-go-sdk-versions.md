# AWS Go SDK Versions

The Terraform AWS Provider relies on the [AWS SDK for Go](https://aws.amazon.com/sdk-for-go/) which is maintained and published by AWS to allow us to safely and securely interact with AWS API's in a consistent fashion. There are two versions of this API, both of which are considered Generally Available and fully supported by AWS at present.

- [AWS SDKs and Tools maintenance policy
](https://docs.aws.amazon.com/sdkref/latest/guide/maint-policy.html)
- [AWS SDKs and Tools version support matrix
](https://docs.aws.amazon.com/sdkref/latest/guide/version-support-matrix.html)

The vast majority of the provider is based on the [AWS Go SDK V1](https://github.com/aws/aws-sdk-go), however the provider allows the use of either.

## Which SDK Version should I use?

Each Terraform provider implementation for an AWS service relies on a service client which in turn is constructed based on a specific SDK version. At present we are slowly increasing our footprint on SDK V2, but are not actively migrating V1 use to V2. The choice of SDK will be as follows:

For new services, you should use [AWS SDK Go V2](https://github.com/aws/aws-sdk-go-v2). We have built a scaffolding tool named [skaff](../skaff/readme.md) which generates new resource or datasource files based on this version of the SDK.

For existing services, you should use whatever version of the SDK that service currently uses. You can determine this by looking at [internal/conns/config.go](https://github.com/hashicorp/terraform-provider-aws/blob/main/internal/conns/config.go) which acts to register the SDK version with the service.

At this time there is no scaffolding support for V1 resources.

## What does the SDK handle?

TODO: retry/auth/

## How do the SDK versions differ?

TODO:



https://aws.github.io/aws-sdk-go-v2/docs/migrating/
