// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tagresource/main.go -UpdateTagsFunc=updateTagsV2
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsOfResource -ServiceTagsSlice -UpdateTags -ParentNotFoundErrCode=ResourceNotFoundException
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -GetTag -ListTags -ListTagsOp=ListTagsOfResource -ServiceTagsSlice -TagsFunc=TagsV2 -KeyValueTagsFunc=keyValueTagsV2 -GetTagsInFunc=getTagsInV2 -SetTagsOutFunc=setTagsOutV2 -ListTagsFunc=listTagsV2 -UpdateTagsFunc=updateTagsV2 -UpdateTags -ParentNotFoundErrCode=ResourceNotFoundException -- tagsv2_gen.go
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package dynamodb
