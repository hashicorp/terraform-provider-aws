# Adding Resource Import Support

Adding import support for Terraform resources will allow existing infrastructure to be managed within Terraform. This type of enhancement generally requires a small to moderate amount of code changes.

Comprehensive code examples and information about resource import support can be found in the [Terraform Plugin SDK v2 documentation](https://www.terraform.io/plugin/sdkv2/resources/import).

- _Resource Code Implementation_: In the resource code (e.g., `internal/service/{service}/{thing}.go`), implementation of `Importer` `State` function. When possible, prefer using [`schema.ImportStatePassthroughContext`](https://www.terraform.io/plugin/sdkv2/resources/import#importer-state-function) as the `Importer` `State` function
- _Resource Acceptance Testing Implementation_: In the resource acceptance testing (e.g., `internal/service/{service}/{thing}_test.go`), implementation of `TestStep`s with `ImportState: true`
- _Resource Documentation Implementation_: In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`), addition of `Import` documentation section at the bottom of the page
