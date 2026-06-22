<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Adding Resource Import Support

Adding import support for Terraform resources will allow existing infrastructure to be managed within Terraform.
This type of enhancement generally requires a small to moderate amount of code changes.
Comprehensive code examples and information about resource import support can be found in the [Terraform Plugin Framework documentation](https://developer.hashicorp.com/terraform/plugin/framework/resources/import).

## With Resource Identity (Preferred)

By default, adding Resource Identity support to a resource type enables importing.
This applies to both Plugin-Framework-based and Plugin-SDK-based resource types.

To enable resource support, see the the [Adding Resource Identity Support Guide](add-resource-identity-support.md).

## Without Resource Identity

- _Resource Code_: In the resource code (e.g., `internal/service/{service}/{thing}.go`),
    - **Plugin Framework (Preferred)** Implement the `ImportState` method on the resource struct. When possible, prefer using the [`resource.ImportStatePassthroughID` function](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/resource#ImportStatePassthroughID).
    - **Plugin SDK V2**: Implement an `Importer` `State` function. When possible, prefer using [`schema.ImportStatePassthroughContext`](https://www.terraform.io/plugin/sdkv2/resources/import#importer-state-function).
- _Resource Acceptance Tests_: In the resource acceptance tests (e.g., `internal/service/{service}/{thing}_test.go`), implement one or more tests containing a `TestStep` with `ImportState: true`.
- _Resource Documentation_: In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`), add an `Import` section at the bottom of the page.
