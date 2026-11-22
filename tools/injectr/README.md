# injectr
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

A utility that maps each Terraform resource and data source to the IAM permissions its handlers need. It combines the AWS SDK
calls found by [`sdkmap`](../sdkmap/README.md) with the operations listed in the
[AWS Service Reference](https://docs.aws.amazon.com/service-authorization/latest/reference/service-reference.html), and writes
`permissions_map.json`: `sdkmap.json` with `iam_permissions` added to every handler.

The high-level workflow is described below:

- Read `sdkmap.json`, written by `sdkmap` beforehand. `injectr` treats it as accurate.
- Fetch the list of AWS Service Reference files and download those that are new or changed since the last run into
  `service-data/`. Delete files for services no longer listed, and record what was fetched in `service-urls-cache.json`.
- Parse the `Operations` of every service reference file and index them by Boto3 client and operation name.
- Map each SDK call in `sdkmap.json`, e.g. `elasticloadbalancingv2.CreateLoadBalancer`, to its Boto3 client, e.g. `elbv2`,
  and look up the IAM actions the operation requires, e.g. `elasticloadbalancing:CreateLoadBalancer`.
- Add the sorted, deduplicated IAM actions to each handler (CRUD, import, `customize_diff` and its nested functions,
  `modify_plan`, tags, and data source `read`) as `iam_permissions`. SDK calls that can't be mapped are listed in
  `unresolved_sdk_calls`, with the reason: `service-has-no-operations`, `operation-not-found`, or `no-iam-actions`.
- Write the results as a JSON and print a summary of resolved and unresolved SDK calls.
- Add an `## IAM Permissions` section to the end of the selected resource and data source docs in
  `website/docs/r` and `website/docs/d`. It lists the permissions needed to plan (`read`, `customize_diff`,
  `modify_plan`, `import`, and tags list) and, for resources, to apply (those in plan plus `create`, `update`, `delete`, and tags
  update). An existing section is updated in place, so re-running never duplicates it.

## Usage

From the repository root, generate `sdkmap.json`, then run `injectr`:

```bash
go run tools/sdkmap/main.go
go -C tools/injectr run .
```

`injectr` is its own Go module, made up of several files, so run it as a package. `go run tools/injectr/main.go` does not
work. Relative paths in the options below are relative to `tools/injectr`.

To skip checking for updated service reference files, e.g. when re-running after a recent sync:

```bash
go -C tools/injectr run . -skip-sync
```

To add IAM permissions to the docs of a service, or a single doc, from the existing `permissions_map.json`:

```bash
go -C tools/injectr run . -skip-sync -skip-generate -insert-service sqs
go -C tools/injectr run . -skip-sync -skip-generate -insert-file website/docs/r/sqs_queue.html.markdown
```

**Options:**

- `-sdkmap FILE` - Path of the `sdkmap` JSON file to read (default: `../../sdkmap.json`)
- `-permissions-map FILE` - Path of the JSON file to write (default: `../../permissions_map.json`)
- `-out DIR` - Directory to download service reference files to (default: `service-data`)
- `-index FILE` - Path of the service reference cache file (default: `service-urls-cache.json`)
- `-skip-sync` - Use the existing service reference files without checking for updates (default: `false`)
- `-skip-generate` - Use the existing `permissions_map.json` without regenerating it (default: `false`)
- `-insert-file FILE` - Doc to add IAM permissions to, relative to `-root`; repeatable
- `-insert-service NAME` - Service package whose docs to add IAM permissions to, e.g. `s3`; repeatable
- `-insert-all` - Add IAM permissions to every resource and data source doc (default: `false`)
- `-dry-run` - Report doc changes without writing them (default: `false`)
- `-root DIR` - Path to the root of the repository (default: `../..`)
