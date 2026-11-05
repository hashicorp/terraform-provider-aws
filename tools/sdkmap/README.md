# sdkmap
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

A static-analysis utility that maps Terraform resource and data source types to AWS SDK service-client operations detected in
their supported service-package handlers.

The high-level workflow is:

- Identify `conns.AWSClient` methods that return AWS SDK service clients and index service packages.
- Find `@SDKResource`, `@SDKDataSource`, `@FrameworkResource`, and `@FrameworkDataSource` factory functions.
- Follow each resource's CRUD, import, and `CustomizeDiff` (SDK V2) or `ModifyPlan` (Framework) handlers to record AWS SDK calls. For data sources, follow only `Read`.
- Include tag-interceptor operations for tagged resources and data sources.
- Write the results as JSON.

## Usage

From the repository root:

```bash
go run tools/sdkmap/main.go
```

**Options:**

- `-root DIR` - Path to the root of the repository (default: `.`)
- `-out FILE` - Path of the JSON file to write (default: `sdkmap.json`)

## Output

The output file is a JSON object with two keys:

| Field | Type | Description |
| --- | --- | --- |
| `resources` | array of objects (resource) | Resources annotated with `@SDKResource` or `@FrameworkResource`. Empty when none are found. |
| `dataSources` | array of objects (data source) | Data sources annotated with `@SDKDataSource` or `@FrameworkDataSource`. Empty when none are found. |

```json
{
  "resources": [],
  "dataSources": []
}
```

### Resource object

| Field | Type | Description |
| --- | --- | --- |
| `service` | string | Service package name. |
| `resource_type` | string | Terraform resource type. |
| `plugin_type` | string | Terraform plugin library: `SDKv2` or `Framework`. |
| `expr` | string | Annotated resource factory expression. |
| `location` | string | Factory declaration location, as a repository-relative `file:line`. |
| `create`, `read`, `update`, `delete` | object (function reference) | CRUD handler. Omitted when not set or not declared in a service package. |
| `customize_diff` | object (function reference) | SDK V2 CustomizeDiff handler. Omitted when not set. |
| `modify_plan` | object (function reference) | Framework `ModifyPlan` method, the CustomizeDiff equivalent. Omitted when not set. |
| `import` | object (function reference) | Custom importer. Omitted when absent or passthrough only. |
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
  "modify_plan": {
    "expr": "",
    "location": "",
    "sdk_calls": []
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

### Data source object

| Field | Type | Description |
| --- | --- | --- |
| `service` | string | Service package name. |
| `data_source_type` | string | Terraform data source type. |
| `plugin_type` | string | Terraform plugin library: `SDKv2` or `Framework`. |
| `expr` | string | Annotated data source factory expression. |
| `location` | string | Factory declaration location, as a repository-relative `file:line`. |
| `read` | object (function reference) | Read handler: the SDK V2 `Read` function or the Framework `Read` method. Omitted when not set or not declared in a service package. |
| `tags` | object (tags) | Tag handlers used by the tags interceptor. Only `list` is set, because the tags interceptor never updates a data source's tags. Omitted for data sources without `@Tags(identifierAttribute=...)`. |

The following value-free JSON shows the complete shape of one data source object. Optional fields are omitted when they do not apply.

```json
{
  "service": "",
  "data_source_type": "",
  "plugin_type": "",
  "expr": "",
  "location": "",
  "read": {
    "expr": "",
    "location": "",
    "sdk_calls": []
  },
  "tags": {
    "list": {
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
| `expr` | string | Function or method expression, e.g. `resourceFooCreate` or `(*fooResource).Create`. |
| `location` | string | Declaration or expression location, as a repository-relative `file:line`. Omitted when unavailable. |
| `sdk_calls` | array of strings | Sorted AWS SDK client methods reachable from the function. |
| `functions` | array of objects (function reference) | Flattened functions combined by a CustomizeDiff combinator. Omitted otherwise. |

### Tags object

| Field | Type | Description |
| --- | --- | --- |
| `list` | object (function reference) | Service package `ListTags` method. Omitted when it is not defined. |
| `update` | object (function reference) | Service package `UpdateTags` method. Omitted when it is not defined, and always omitted for data sources. |

## Limitations

- S3 calls made through the AWS SDK's [`feature/s3/manager`](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/manager) package aren't tracked. For example, the `aws_s3_object` and `aws_s3_bucket_object` uploads (`s3.PutObject` and the multipart upload operations) are missing from their `create` and `update` entries.
- Tag listers and updaters shared by several resource types, e.g. the IAM and S3 `ListTags` and `UpdateTags`, are narrowed to a resource's or data source's type only where they, or a function they pass it to, use `switch resourceType` with string literal cases. Any other branching on the resource type is analyzed as a whole, so the `tags` entry includes the SDK calls for every resource type that shares them.
- Only service packages are analyzed. When an AWS SDK client is passed to a function in any other package, e.g. an SDK waiter, an SDK presign client, or the S3 transfer manager, the operations it calls aren't tracked, and sdkmap logs a warning naming the resource's or data source's function and the function the client is passed to.
- Calls through an interface, including one embedded in a Framework resource or data source type, include every implementation in the package. For example, `aws_vpc_security_group_ingress_rule` and `aws_vpc_security_group_egress_rule` include each other's SDK calls.
- Framework ephemeral resources, actions, list resources, attribute-level `PlanModifiers`, and resource and data source methods other than those described above (e.g. `ValidateConfig`) aren't analyzed.
