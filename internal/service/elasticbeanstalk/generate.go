// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -KVTValues -ListTags -ListTagsOutTagsElem=ResourceTags -ServiceTagsSlice -TagOp=UpdateTagsForResource -TagInTagsElem=TagsToAdd -UntagOp=UpdateTagsForResource -UntagInTagsElem=TagsToRemove -UpdateTags
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=DescribeEnvironments
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package elasticbeanstalk
