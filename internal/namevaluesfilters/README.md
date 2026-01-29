<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# namevaluesfilters

The `namevaluesfilters` package is designed to provide a consistent interface for handling AWS resource filtering with AWS SDK for Go v2.

This package implements a single `NameValuesFilters` type, which covers all filter handling logic, such as merging filters, via functions on the single type. The underlying implementation is compatible with Go operations such as `len()`.
