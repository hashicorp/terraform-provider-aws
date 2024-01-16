// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTags -ListTagsInIDElem=ARN -ListTagsOutTagsElem=TagList -ServiceTagsSlice -TagOp=AddTags -TagInIDElem=ARN -TagInTagsElem=TagList -UntagOp=RemoveTags -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package opensearch
