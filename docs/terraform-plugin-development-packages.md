# Terraform Plugin Development Packages

The Terraform AWS Provider is constructed with HashiCorp-maintained packages for building plugins.
Most existing resources are implemented with [Terraform Plugin SDK V2](https://developer.hashicorp.com/terraform/plugin/sdkv2), while newer resources use [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).
A thorough comparison of the packages can be found [here](https://developer.hashicorp.com/terraform/plugin/framework-benefits).

## Which Plugin Version Should I Use?

At this time net-new contributions are **required to use Terraform Plugin Framework**.
Enhancements or bug fixes to existing Plugin SDKv2 based resources do not require migration.
The AWS Provider is [muxed](https://developer.hashicorp.com/terraform/plugin/framework/migrating/mux) to allow existing Terraform Plugin SDK V2 resources and data sources to remain alongside newer Plugin Framework based resources.

[`skaff`](skaff.md) will generate Plugin Framework based resources by default, though for exceptional cases the optional `-p`/`--plugin-sdkv2` flag can be used.
Where applicable, the contributor guide has been updated to include examples with both Terraform Plugin Framework and Terraform Plugin SDK V2.
