# AWS Go SDK Versions

The Terraform AWS Provider relies on the [AWS SDK for Go](https://aws.amazon.com/sdk-for-go/) which is maintained and published by AWS to allow us to safely and securely interact with AWS APIs consistently.
There are two versions of this API, both of which are considered Generally Available and fully supported by AWS at present.

- [AWS SDKs and Tools maintenance policy](https://docs.aws.amazon.com/sdkref/latest/guide/maint-policy.html)
- [AWS SDKs and Tools version support matrix](https://docs.aws.amazon.com/sdkref/latest/guide/version-support-matrix.html)

Each Terraform provider implementation for an AWS service relies on a service client, which in turn is constructed based on a specific SDK version.
While the vast majority of the provider is based on the [AWS SDK for Go v1](https://github.com/aws/aws-sdk-go), the provider also allows the use of the [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2).

## Which SDK Version should I use?

At this time **all net-new services are required to use AWS Go SDK v2**.
Enhancements or bug fixes to existing AWS Go SDK v1 based services do not require migration.
You can determine the SDK version a given service uses by looking at the `import` section in the service's Go files.
Where applicable, the contributor guide has been updated to include examples with both AWS Go SDK v1 and v2.

We are beginning to migrate core services written with v1 onto v2, and adoption will continue to increase over time.
AWS has a [migration guide](https://aws.github.io/aws-sdk-go-v2/docs/migrating/) that details the differences between the versions of the SDK.

## What does the SDK handle?

The AWS SDKs handle calling the various web service interfaces for AWS services.
In addition to encoding and decoding the Go structures in the correct JSON or XML payloads,
the SDKs handle authentication, request logging, and retrying requests.

The various language SDKs and the AWS CLI share a consistent configuration interface,
using environment variables and shared configuration and credentials files.

The AWS SDKs also automatically retry several common failure cases, such as network errors.

## How do the SDK versions differ?

The AWS SDK for Go v1.0.0 was released in late 2015, when the current version of Go was v1.5.
The Go language has evolved significantly since then.
Many currently recommended practices were not possible at that time,
including the use of the `context` package, introduced in Go v1.7,
and error wrapping, introduced in Go v1.13.

The AWS SDK for Go v2 uses a modern Go style
and has also been modularized so that individual services are packaged and imported separately.

For details on the specific changes to the AWS SDK for Go v2,
see [Migrating to the AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/docs/migrating/),
especially the [Service Clients](https://aws.github.io/aws-sdk-go-v2/docs/migrating/#service-clients) section.
