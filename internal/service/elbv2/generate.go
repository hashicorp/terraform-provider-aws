// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceArns -ListTagsInIDNeedSlice=yes      -ListTagsOutTagsElem=TagDescriptions[0].Tags -ServiceTagsSlice -TagOp=AddTags -TagInIDElem=ResourceArns -TagInIDNeedSlice=yes      -UntagOp=RemoveTags -UpdateTags -CreateTags -TagsFunc=tags   -KeyValueTagsFunc=keyValueTags
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceArns -ListTagsInIDNeedValueSlice=yes -ListTagsOutTagsElem=TagDescriptions[0].Tags -ServiceTagsSlice -TagOp=AddTags -TagInIDElem=ResourceArns -TagInIDNeedValueSlice=yes -UntagOp=RemoveTags -UpdateTags -CreateTags -TagsFunc=tagsV2 -KeyValueTagsFunc=keyValueTagsV2 -ListTagsFunc=listTagsV2 -GetTagsInFunc=getTagsInV2 -SetTagsOutFunc=setTagsOutV2 -UpdateTagsFunc=updateTagsV2 -CreateTagsFunc=createTagsV2 -AWSSDKVersion=2 -KVTValues -- tagsv2_gen.go
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package elbv2
