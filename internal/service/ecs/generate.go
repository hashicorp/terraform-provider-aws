// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=DescribeCapacityProviders
//go:generate go run ../../generate/tagresource/main.go
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -GetTag -ListTags -ServiceTagsSlice -UpdateTags -CreateTags -ParentNotFoundErrCode=InvalidParameterException "-ParentNotFoundErrMsg=The specified cluster is inactive. Specify an active cluster and try again."
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package ecs
