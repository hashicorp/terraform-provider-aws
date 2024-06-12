// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ServiceTagsSlice -AWSSDKVersion=2 -ListTags -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tags/main.go -ServiceTagsSlice -AWSSDKVersion=2 -ListTags -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package appfabric
