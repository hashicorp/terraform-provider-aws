# injectr
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

A utility that maps each Terraform resource and data source to the AWS SDK calls and IAM permissions its handlers need,
and documents them. It combines the AWS SDK calls found by its [`sdkmap`](sdkmap/README.md) package with the operations
listed in the [AWS Service Reference](https://docs.aws.amazon.com/service-authorization/latest/reference/service-reference.html).

Workflow:

- Run `sdkmap` over the service packages under `-root` to find every resource and data source and the AWS SDK calls its handlers make. See the [`sdkmap` README](sdkmap/README.md) for its limitations.
- Fetch the list of AWS Service Reference files and download those that are new or changed since the last run into `service-data/`. Delete files for services no longer listed, and record what was fetched in `service-urls-cache.json`.
- Parse the `Operations` of every service reference file and index them by Boto3 client and operation name.
- Map each SDK call, e.g. `elasticloadbalancingv2.CreateLoadBalancer`, to its Boto3 client, e.g. `elbv2`, and look up the IAM actions the operation requires, e.g. `elasticloadbalancing:CreateLoadBalancer`.
- Add the sorted, deduplicated IAM actions to each handler (CRUD, import, `customize_diff` and its nested functions, `modify_plan`, tags, and data source `read`) as `iam_permissions`. SDK calls that can't be mapped are listed in `unresolved_sdk_calls`, with the reason: `service-has-no-operations`, `operation-not-found`, or `no-iam-actions`.
- Write `permissions_map.json`, every resource and data source with each handler's `sdk_calls` and `iam_permissions`, and print a summary of resolved and unresolved SDK calls.
- Add an `## IAM Permissions` section to the end of the selected resource and data source docs in `website/docs/r` and `website/docs/d`. It lists the permissions needed to plan (`read`, `customize_diff`, `modify_plan`, `import`, and tags list) and, for resources, to apply (those in plan plus `create`, `update`, `delete`, and tags update). An existing section is updated in place, so re-running does not duplicate it.

## Usage

`injectr` is its own Go module; run it with `go -C tools/injectr run .`. Each run writes `permissions_map.json`. For example, to also insert IAM permissions for a service, e.g. `sns`, from the repository root:

```bash
go -C tools/injectr run . -insert-service sns
```

To skip checking for updated service reference files, e.g. when re-running after a recent sync:

```bash
go -C tools/injectr run . -skip-sync -insert-service sqs,sns
go -C tools/injectr run . -skip-sync -insert-file website/docs/r/sqs_queue.html.markdown,website/docs/d/sqs_queue.html.markdown
```

## Options

### Paths

Relative paths are relative to `tools/injectr`.

- `-root DIR` - Path to the root of the repository, whose service packages are analyzed and whose docs are updated (default: `../..`)
- `-permissions-map FILE` - Path of the permissions map JSON file to write (default: `../../permissions_map.json`)
- `-out DIR` - Directory to download service reference files to (default: `service-data`)
- `-index FILE` - Path of the service reference cache file (default: `service-urls-cache.json`)

### Service reference sync

- `-skip-sync` - Use the existing service reference files without checking for updates (default: `false`)

### Doc insertion

Docs are only changed when at least one of `-insert-file`, `-insert-service`, or `-insert-all` is set.

- `-insert-file FILES` - Comma-separated docs to add IAM permissions to, relative to `-root`
- `-insert-service NAMES` - Comma-separated service packages to add IAM permissions to, e.g. `s3,sqs`
- `-insert-all` - Add IAM permissions to every resource and data source doc (default: `false`)
- `-dry-run` - Report doc changes without writing them. `permissions_map.json` is still written (default: `false`)

## Output

`permissions_map.json` is a JSON object with two keys:

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
| `customize_diff` | object (function reference) | SDKv2 CustomizeDiff handler. Omitted when not set. |
| `modify_plan` | object (function reference) | Framework `ModifyPlan` method, the CustomizeDiff equivalent. Omitted when not set. |
| `import` | object (function reference) | Custom importer. Omitted when absent or passthrough only. |
| `tags` | object (tags) | Tag handlers used by the tags interceptor. Omitted for resources without `@Tags(identifierAttribute=...)`. |

The following value-free JSON shows the shape of one resource object. Optional fields are omitted when they do not apply. Function references are abbreviated; see [Function reference](#function-reference) for their complete shape.

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
    "sdk_calls": [],
    "iam_permissions": []
  },
  "read": {
    "expr": "",
    "location": "",
    "sdk_calls": [],
    "iam_permissions": []
  },
  "update": {
    "expr": "",
    "location": "",
    "sdk_calls": [],
    "iam_permissions": []
  },
  "delete": {
    "expr": "",
    "location": "",
    "sdk_calls": [],
    "iam_permissions": []
  },
  "customize_diff": {
    "expr": "",
    "location": "",
    "sdk_calls": [],
    "iam_permissions": [],
    "functions": [
      {
        "expr": "",
        "location": "",
        "sdk_calls": [],
        "iam_permissions": []
      }
    ]
  },
  "modify_plan": {
    "expr": "",
    "location": "",
    "sdk_calls": [],
    "iam_permissions": []
  },
  "import": {
    "expr": "",
    "location": "",
    "sdk_calls": [],
    "iam_permissions": []
  },
  "tags": {
    "list": {
      "expr": "",
      "location": "",
      "sdk_calls": [],
      "iam_permissions": []
    },
    "update": {
      "expr": "",
      "location": "",
      "sdk_calls": [],
      "iam_permissions": []
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
| `read` | object (function reference) | Read handler: the SDKv2 `Read` function or the Framework `Read` method. Omitted when not set or not declared in a service package. |
| `tags` | object (tags) | Tag handlers used by the tags interceptor. Only `list` is set. Omitted for data sources without `@Tags(identifierAttribute=...)`. |

The following value-free JSON shows the shape of one data source object, with function references abbreviated as above.

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
    "sdk_calls": [],
    "iam_permissions": []
  },
  "tags": {
    "list": {
      "expr": "",
      "location": "",
      "sdk_calls": [],
      "iam_permissions": []
    }
  }
}
```

### Function reference

Each handler field, and each field in a tags object, is a function reference. `sdkmap` sets `sdk_calls`; `permissions` sets `iam_permissions` and `unresolved_sdk_calls`.

| Field | Type | Description |
| --- | --- | --- |
| `expr` | string | Function or method expression, e.g. `resourceFooCreate` or `(*fooResource).Create`. |
| `location` | string | Declaration or expression location, as a repository-relative `file:line`. Omitted when unavailable. |
| `sdk_calls` | array of strings | Sorted AWS SDK client methods reachable from the function, e.g. `s3.PutObjectAcl`. |
| `iam_permissions` | array of strings | Sorted, deduplicated IAM actions required by `sdk_calls`, e.g. `s3:PutObjectAcl`. |
| `unresolved_sdk_calls` | array of objects | SDK calls that could not be mapped to IAM actions, each with its `sdk_call` and `reason`: `service-has-no-operations`, `operation-not-found`, or `no-iam-actions`. Omitted when every call is resolved. |
| `functions` | array of objects (function reference) | Flattened functions combined by a CustomizeDiff combinator. Omitted otherwise. |

```json
{
  "expr": "",
  "location": "",
  "sdk_calls": [],
  "iam_permissions": [],
  "unresolved_sdk_calls": [
    {
      "sdk_call": "",
      "reason": ""
    }
  ],
  "functions": []
}
```

### Tags object

| Field | Type | Description |
| --- | --- | --- |
| `list` | object (function reference) | Service package `ListTags` method. Omitted when it is not defined. |
| `update` | object (function reference) | Service package `UpdateTags` method. Omitted when it is not defined, and always omitted for data sources. |

## Packages

- `main.go` - The CLI: runs `sdkmap`, `serviceref`, `permissions`, then `docs`.
- `sdkmap` - Analyzes the service packages and finds the AWS SDK calls made by each resource and data source. See its [README](sdkmap/README.md).
- `serviceref` - Downloads the AWS Service Reference files and parses them.
- `permissions` - Maps each SDK call to its IAM actions using the service reference data, and writes `permissions_map.json`.
- `docs` - Adds the `## IAM Permissions` section to resource and data source docs.
- `model` - The `permissions_map.json` types: built by `sdkmap`, extended by `permissions`, and read by `docs`.

Test fixtures are in each package's `testdata` directory. The `sdkmap` tests use a miniature provider root and golden files in `sdkmap/testdata`; see its [README](sdkmap/testdata/README.md). The `permissions` tests use the service reference fixtures in `serviceref/testdata` and the SDK call fixture `permissions/testdata/sdkmap_input.json`. The `docs` tests use `permissions/testdata/permissions_map.json`. To regenerate that file, run `go test ./permissions -run TestGeneratePermissionsMap -update` from `tools/injectr`.
