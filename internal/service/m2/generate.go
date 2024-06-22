// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tags/main.go -ListTags -ServiceTagsMap -UpdateTags -KVTValues -AWSSDKVersion=2 -SkipTypesImp
// ONLY generate directives and package declaration! Do not add anything else to this file

package m2
