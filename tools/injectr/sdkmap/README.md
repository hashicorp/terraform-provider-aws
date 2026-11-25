# sdkmap
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

Static analysis, for [`injectr`](../README.md), that finds the AWS SDK calls made by each Terraform resource's and data
source's handlers.

Workflow:

- Identify `conns.AWSClient` methods that return AWS SDK service clients and index service packages.
- Find `@SDKResource`, `@SDKDataSource`, `@FrameworkResource`, and `@FrameworkDataSource` factory functions.
- Follow each resource's CRUD, import, and `CustomizeDiff` (SDKv2) or `ModifyPlan` (Framework) handlers to record AWS SDK calls. For data sources, follow only `Read`.
- Include tag-interceptor operations for tagged resources and data sources.
- Return the results as a [`model.Map`](../model/model.go), shared with the rest of `injectr`.

## Usage

`sdkmap` has no command and writes no files: `injectr` calls `Generate` with its `-root` and adds IAM permissions to the
result. See the [`injectr` README](../README.md#usage). Analysis warnings are logged to standard error.

To run its tests, from `tools/injectr`:

```bash
go test ./sdkmap
```

## Output

`Generate` returns a [`model.Map`](../model/model.go), which `injectr` writes as `permissions_map.json`. See the
[`injectr` README](../README.md#output) for its format.

## Limitations

- S3 calls made through the AWS SDK's [`feature/s3/manager`](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/manager) package aren't tracked. For example, the `aws_s3_object` and `aws_s3_bucket_object` uploads (`s3.PutObject` and the multipart upload operations) are missing from their `create` and `update` entries.
- Tag listers and updaters shared by several resource types, e.g. the IAM and S3 `ListTags` and `UpdateTags`, are narrowed to a resource's or data source's type only where they, or a function they pass it to, use `switch resourceType` with string literal cases. Any other branching on the resource type is analyzed as a whole, so the `tags` entry includes the SDK calls for every resource type that shares them.
- Only service packages are analyzed. When an AWS SDK client is passed to a function in any other package, e.g. an SDK waiter, an SDK presign client, or the S3 transfer manager, the operations it calls aren't tracked; a warning is logged.
- Calls through an interface, including one embedded in a Framework resource or data source type, include every implementation in the package. For example, `aws_vpc_security_group_ingress_rule` and `aws_vpc_security_group_egress_rule` include each other's SDK calls.
- Framework ephemeral resources, actions, list resources, attribute-level `PlanModifiers`, and resource and data source methods other than those described above (e.g. `ValidateConfig`) aren't analyzed.
