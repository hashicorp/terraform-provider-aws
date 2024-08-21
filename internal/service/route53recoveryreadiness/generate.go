// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsOp=ListTagsForResources -UpdateTags -CreateTags -ServiceTagsMap -KVTValues -SkipTypesImp
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package route53recoveryreadiness
