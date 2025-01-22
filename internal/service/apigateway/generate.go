// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -ListOps=GetAuthorizers,GetDomainNameAccessAssociations -Paginator=Position
//go:generate go run ../../generate/tags/main.go -ServiceTagsMap -UpdateTags -KVTValues -ListTags -ListTagsOp=GetTags
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package apigateway
