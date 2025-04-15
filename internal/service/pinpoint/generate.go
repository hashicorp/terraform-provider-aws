// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOutTagsElem=TagsModel.Tags -ServiceTagsMap "-TagInCustomVal=&awstypes.TagsModel{Tags: svcTags(updatedTags.IgnoreAWS())}" -TagInTagsElem=TagsModel -UpdateTags -KVTValues
//go:generate go run ../../generate/listpages/main.go -ListOps=GetApps -OutputPaginator=ApplicationsResponse.NextToken -InputPaginator=Token
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package pinpoint
