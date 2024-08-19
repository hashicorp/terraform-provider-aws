// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsOp=ListTags -ListTagsOpPaginated -ListTagsInIDElem=ResourceIdList --ListTagsInIDNeedValueSlice=yes -ListTagsOutTagsElem=ResourceTagList[0].TagsList -ServiceTagsSlice -TagOp=AddTags -TagInIDElem=ResourceId -TagInTagsElem=TagsList -UntagOp=RemoveTags -UntagInNeedTagType -UntagInTagsElem=TagsList -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package cloudtrail
