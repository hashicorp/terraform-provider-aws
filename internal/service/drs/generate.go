// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsInIDElem=ResourceArn -ServiceTagsMap -SkipTypesImp -KVTValues -TagOp=TagResource -TagInIDElem=ResourceArn -UntagOp=UntagResource -CreateTags -ListTags -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package drs
