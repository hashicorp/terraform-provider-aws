// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=ResourceName -ListTagsOutTagsElem=TagList -ServiceTagsSlice -TagOp=AddTagsToResource -TagInIDElem=ResourceName -UntagOp=RemoveTagsFromResource -UpdateTags
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ServiceTagsSlice -TagsFunc=TagsV2 -KeyValueTagsFunc=keyValueTagsV2 -GetTagsInFunc=getTagsInV2 -SetTagsOutFunc=setTagsOutV2 -SkipAWSServiceImp -- tagsv2_gen.go
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package rds
