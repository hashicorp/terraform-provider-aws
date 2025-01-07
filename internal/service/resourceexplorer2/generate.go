// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -TagInIDElem=ResourceArn -ListTags -ListTagsInIDElem=ResourceArn -ServiceTagsMap -UpdateTags -UntagInTagsElem=TagKeys -KVTValues
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package resourceexplorer2
