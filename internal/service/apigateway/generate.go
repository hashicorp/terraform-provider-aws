// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -ListOps=GetAuthorizers -Paginator=Position -AWSSDKVersion=2
//go:generate go run ../../generate/tags/main.go -ServiceTagsMap -UpdateTags -AWSSDKVersion=2 -KVTValues -SkipTypesImp -ListTags -ListTagsOp=GetTags
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package apigateway
