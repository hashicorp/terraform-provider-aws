// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ServiceTagsMap -UpdateTags
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -TagsFunc=TagsV2 -ListTagsFunc=ListTagsV2  -KeyValueTagsFunc=keyValueTagsV2 -GetTagsInFunc=getTagsInV2 -SetTagsOutFunc=setTagsOutV2 -UpdateTagsFunc=updateTagsV2 -ListTags -ServiceTagsMap -UpdateTags -KVTValues -SkipTypesImp -- tagsv2_gen.go
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package kafka
