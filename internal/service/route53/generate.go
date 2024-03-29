// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -ListOps=ListTrafficPolicies -Paginator=TrafficPolicyIdMarker list_traffic_policies_pages_gen.go
//go:generate go run ../../generate/listpages/main.go -ListOps=ListTrafficPolicyVersions -Paginator=TrafficPolicyVersionMarker list_traffic_policy_versions_pages_gen.go
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=ResourceId -ListTagsOutTagsElem=ResourceTagSet.Tags -ServiceTagsSlice -TagOp=ChangeTagsForResource -TagInIDElem=ResourceId -TagInTagsElem=AddTags -TagResTypeElem=ResourceType -UntagOp=ChangeTagsForResource -UntagInTagsElem=RemoveTagKeys -UpdateTags -CreateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package route53
