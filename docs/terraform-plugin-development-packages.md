# Terraform Plugin Development Packages

The Terraform AWS Provider is constructed with HashiCorp-maintained packages for building plugins.
All new resources use [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework), however, there are many existing resources implemented with [Terraform Plugin SDK V2](https://developer.hashicorp.com/terraform/plugin/sdkv2).
A thorough comparison of the packages can be found [here](https://developer.hashicorp.com/terraform/plugin/framework-benefits).

## Which Plugin Version Should I Use?

All net-new contributions are **required to use Terraform Plugin Framework**.
Maintainers will migrate resources to Terraform Plugin Framework during the review process if necessary, but efforts toward submitting the initial implementation in Terraform Plugin Framework are greatly appreciated.

Enhancements or bug fixes to existing Plugin SDKV2 based resources **do not require migration**.
The AWS Provider is [muxed](https://developer.hashicorp.com/terraform/plugin/framework/migrating/mux) to allow existing Terraform Plugin SDK V2 resources and data sources to remain alongside newer Plugin Framework based resources.

[`skaff`](skaff.md) will generate Plugin Framework based resources by default.
Where applicable, the contributor guide includes code examples using both Terraform Plugin Framework and Terraform Plugin SDK V2.
