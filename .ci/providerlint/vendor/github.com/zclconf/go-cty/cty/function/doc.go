// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package function builds on the functionality of cty by modeling functions
// that operate on cty Values.
//
// Functions are, at their core, Go anonymous functions. However, this package
// wraps around them utility functions for parameter type checking, etc.
package function
