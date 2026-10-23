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

## Output

The output file is a JSON array of resource objects.

### Resource object

| Field | Type | Description |
| --- | --- | --- |
| `service` | string | Service package name. |
| `resource_type` | string | Terraform resource type. |
| `plugin_type` | string | Terraform plugin library that implements the resource. |
| `expr` | string | Annotated resource factory expression. |
| `location` | string | Factory declaration location, as a repository-relative `file:line`. |
| `create`, `read`, `update`, `delete` | object (function reference) | CRUD handler. Omitted when the handler is not set. |
| `customize_diff` | object (function reference) | CustomizeDiff handler. Omitted when not set. |
| `import` | object (function reference) | Custom importer. Omitted when no custom importer is used. |
| `tags` | object (tags) | Tag handlers used by the tags interceptor. Omitted for resources without `@Tags(identifierAttribute=...)`. |

The following value-free JSON shows the complete shape of one resource object. Optional fields are omitted when they do not apply.

```json
{
  "service": "",
  "resource_type": "",
  "plugin_type": "",
  "expr": "",
  "location": "",
  "create": {
    "expr": "",
    "location": "",
    "sdk_calls": []
  },
  "read": {
    "expr": "",
    "location": "",
    "sdk_calls": []
  },
  "update": {
    "expr": "",
    "location": "",
    "sdk_calls": []
  },
  "delete": {
    "expr": "",
    "location": "",
    "sdk_calls": []
  },
  "customize_diff": {
    "expr": "",
    "location": "",
    "sdk_calls": [],
    "functions": [
      {
        "expr": "",
        "location": "",
        "sdk_calls": []
      }
    ]
  },
  "import": {
    "expr": "",
    "location": "",
    "sdk_calls": []
  },
  "tags": {
    "list": {
      "expr": "",
      "location": "",
      "sdk_calls": []
    },
    "update": {
      "expr": "",
      "location": "",
      "sdk_calls": []
    }
  }
}
```

### Function reference

Each handler field, and each field in a tags object, is a function reference.

| Field | Type | Description |
| --- | --- | --- |
| `expr` | string | Function expression from the resource or service package. |
| `location` | string | Declaration or expression location, as a repository-relative `file:line`. Omitted when unavailable. |
| `sdk_calls` | array of strings | Sorted AWS SDK client methods reachable from the function. |
| `functions` | array of objects (function reference) | Flattened functions combined by a CustomizeDiff combinator. Omitted otherwise. |

### Tags object

| Field | Type | Description |
| --- | --- | --- |
| `list` | object (function reference) | Service package `ListTags` method. Omitted when it is not defined. |
| `update` | object (function reference) | Service package `UpdateTags` method. Omitted when it is not defined. |

## Limitations

- S3 calls made through the AWS SDK's [`feature/s3/manager`](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/manager) package aren't tracked. For example, the `aws_s3_object` and `aws_s3_bucket_object` uploads (`s3.PutObject` and the multipart upload operations) are missing from their `create` and `update` entries.
- Tag listers and updaters shared by several resource types, e.g. the IAM and S3 `ListTags` and `UpdateTags`, are narrowed to a resource's type only where they, or a function they pass it to, use `switch resourceType` with string literal cases. Any other branching on the resource type is analyzed as a whole, so the resource's `tags` entry includes the SDK calls for every resource type that shares them.
- Only service packages are analyzed. When an AWS SDK client is passed to a function in any other package, e.g. an SDK waiter, an SDK presign client, or the S3 transfer manager, the operations it calls aren't tracked, and sdkmap logs a warning naming the resource's function and the function the client is passed to.
