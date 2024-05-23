// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ServiceTagsMap -UpdateTags -CreateTags -KVTValues -SkipTypesImp
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsOp=ListTagsLogGroup -ListTagsInIDElem=LogGroupName -ListTagsFunc=listLogGroupTags -TagOp=TagLogGroup -TagInIDElem=LogGroupName -UntagOp=UntagLogGroup -UntagInTagsElem=Tags -UpdateTags -UpdateTagsFunc=updateLogGroupTags -SkipTypesImp -- log_group_tags_gen.go
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package logs
