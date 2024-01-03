// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -KVTValues -SkipTypesImp -AWSSDKVersion=2 -ListTags -ListTagsInIDElem=ResourceArn -ListTagsOutTagsElem=Tags -ServiceTagsMap -TagOp=TagResource -TagInIDElem=ResourceArn -UntagOp=UntagResource -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package docdbelastic
