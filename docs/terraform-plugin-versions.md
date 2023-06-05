# Terraform Plugin Versions

The Terraform AWS Provider is constructed with HashiCorp-maintained packages for building plugins. Most existing resources are implemented with [Terraform Plugin SDKv2](https://developer.hashicorp.com/terraform/plugin/sdkv2), while newer resources may use [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework). A thorough comparison of the packages can be found [here](https://developer.hashicorp.com/terraform/plugin/framework-benefits).

At this time community contributions in either package will be accepted. The AWS Provider is [muxed](https://developer.hashicorp.com/terraform/plugin/framework/migrating/mux) to allow resources and data sources implemented in both packages. As AWS Provider tooling around Plugin Framework (and the library itself) matures, we will being requiring all net-new resources be implemented with it. [`skaff`](skaff.md) currently supports generating Plugin Framework based resources using the optional `-p`/`--plugin-framework` flag. Factors to consider when choosing between packages are:

1. What other resources in a given service use
2. Level of comfort with the new idioms introduced in Plugin Framework
3. [Advantages](https://developer.hashicorp.com/terraform/plugin/framework-benefits#plugin-framework-benefits) Plugin Framework may afford over Plugin SDKv2 (improved null handling, plan modifications, etc.)
