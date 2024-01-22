# namevaluesfiltersv2

The `namevaluesfiltersv2` package is designed to provide a consistent interface for handling AWS resource filtering for AWS SDK for Go v2.

This package implements a single `NameValuesFilters` type, which covers all filter handling logic, such as merging filters, via functions on the single type. The underlying implementation is compatible with Go operations such as `len()`.

Full documentation for this package can be found on [GoDoc](https://godoc.org/github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfiltersv2).

Many AWS Go SDK v2 services that support resource filtering have their service-specific Go type conversion functions to and from `NameValuesFilters` code generated. Converting from `NameValuesFilters` to AWS Go SDK v2 types is done via `{SERVICE}Filters()` functions on the type. For more information about this code generation, see the [`generators/servicefilters` README](generators/servicefilters/README.md).

Any filtering functions that cannot be generated should be hand implemented in a service-specific source file and follow the format of similar generated code wherever possible. The first line of the source file should be `// +build !generate`. This prevents the file's inclusion during the code generation phase.

## Code Structure

```text
internal/generate/namevaluesfiltersv2
├── generators
│   └── servicefilters (generates service_filters_gen.go)
├── name_values_filters_test.go (unit tests for core logic)
├── name_values_filters.go (core logic)
├── service_generation_customizations.go (shared AWS Go SDK service customizations for generators)
├── service_filters_gen.go (generated AWS Go SDK service conversion functions)
└── <service name>_filters.go (any service-specific functions that cannot be generated)
```
