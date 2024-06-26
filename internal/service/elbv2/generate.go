// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceArns -ListTagsInIDNeedValueSlice=yes -ListTagsOutTagsElem=TagDescriptions[0].Tags -ServiceTagsSlice -TagOp=AddTags -TagInIDElem=ResourceArns -TagInIDNeedValueSlice=yes -UntagOp=RemoveTags -UpdateTags -CreateTags -AWSSDKVersion=2 -KVTValues
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go -AWSSDKVersion=2
// ONLY generate directives and package declaration! Do not add anything else to this file.

package elbv2
