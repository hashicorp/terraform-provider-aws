// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeQueryDefinitions,DescribeResourcePolicies
//go:generate go run ../../generate/tags/main.go -ListTags -ServiceTagsMap -UpdateTags -CreateTags
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsLogGroup -ListTagsInIDElem=LogGroupName -ListTagsFunc=listLogGroupTags -TagOp=TagLogGroup -TagInIDElem=LogGroupName -UntagOp=UntagLogGroup -UntagInTagsElem=Tags -UpdateTags -UpdateTagsFunc=updateLogGroupTags -- log_group_tags_gen.go
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package logs
