// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsOp=DescribeTags -ListTagsOpPaginated -ListTagsInIDElem=FileSystemId -ServiceTagsSlice -TagInIDElem=ResourceId -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package efs
