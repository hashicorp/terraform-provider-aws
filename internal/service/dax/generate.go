// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTags -ListTagsInIDElem=ResourceName -ServiceTagsSlice -TagInIDElem=ResourceName -UpdateTags
//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeClusters
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package dax
