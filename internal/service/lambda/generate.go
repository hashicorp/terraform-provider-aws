// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ServiceTagsMap -TagInIDElem=Resource -UpdateTags -ListTags -ListTagsInIDElem=Resource -ListTagsOp=ListTags -AWSSDKVersion=2 -KVTValues -SkipTypesImp
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package lambda
