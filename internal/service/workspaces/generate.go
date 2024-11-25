// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeConnectionAliases,DescribeIpGroups,DescribeWorkspaceImages
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceId -ListTagsOutTagsElem=TagList -ServiceTagsSlice -TagOp=CreateTags -TagInIDElem=ResourceId -UntagOp=DeleteTags -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package workspaces
