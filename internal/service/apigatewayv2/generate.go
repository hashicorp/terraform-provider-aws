// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -ListOps=GetApis,GetApiMappings,GetDomainNames,GetIntegrations,GetRoutes,GetStages,GetVpcLinks
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=GetTags -ServiceTagsMap -UpdateTags -KVTValues
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package apigatewayv2
