// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ServiceTagsMap -UpdateTags -ListTags
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -KVTValues -KeyValueTagsFunc=keyValueTagsV2 -GetTagsInFunc=getTagsInV2 -SetTagsOutFunc=setTagsOutV2 -ServiceTagsMap -SkipTypesImp -SkipAWSImp -TagsFunc=TagsV2 -- tagsv2_gen.go
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package location
