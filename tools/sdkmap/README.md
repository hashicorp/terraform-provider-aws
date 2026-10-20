# sdkmap
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

## Usage

From the repository root:

```bash
go run tools/sdkmap/main.go
```

**Options:**

- `-root DIR` - Path to the root of the repository (default: `.`)
- `-out FILE` - Path of the JSON file to write (default: `sdkmap.json`)

## Limitations

- S3 calls made through the AWS SDK's [`feature/s3/manager`](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/manager) package aren't tracked. For example, the `aws_s3_object` and `aws_s3_bucket_object` uploads (`s3.PutObject` and the multipart upload operations) are missing from their `create` and `update` entries.
- Tag listers and updaters that share one entry point for several resource types, e.g. the IAM and S3 `ListTags` and `UpdateTags`, are analyzed as a whole. Each resource's `tags` entry includes the SDK calls for every resource type that shares them, so the permissions it implies are broader than the resource needs.
- Only service packages are analyzed. When an AWS SDK client is passed to a function in any other package, e.g. an SDK waiter, an SDK presign client, or the S3 transfer manager, the operations it calls aren't tracked, and sdkmap logs a warning naming the resource's function and the function the client is passed to.
