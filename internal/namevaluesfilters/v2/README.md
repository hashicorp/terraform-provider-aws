# namevaluesfiltersv2

The `namevaluesfilters/v2` package is designed to provide a consistent interface for handling AWS resource filtering with AWS SDK for Go v2.

This package implements a single `NameValuesFilters` type, which covers all filter handling logic, such as merging filters, via functions on the single type. The underlying implementation is compatible with Go operations such as `len()`.
