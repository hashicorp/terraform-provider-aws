// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../generate/servicepackages/main.go -ServicePackageRoot ../service -- service_packages_gen_test.go
//go:generate go run ../generate/sweeperregistration/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package sweep_test
