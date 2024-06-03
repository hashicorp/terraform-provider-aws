// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListHostedZonesByVPC,ListVPCAssociationAuthorizations
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListTrafficPolicies -Paginator=TrafficPolicyIdMarker -- list_traffic_policies_pages_gen.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListTrafficPolicyVersions -Paginator=TrafficPolicyVersionMarker -- list_traffic_policy_versions_pages_gen.go
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsInIDElem=ResourceId -ListTagsOutTagsElem=ResourceTagSet.Tags -ServiceTagsSlice -TagOp=ChangeTagsForResource -TagInIDElem=ResourceId -TagInTagsElem=AddTags -TagResTypeElem=ResourceType -TagResTypeElemType=TagResourceType -UntagOp=ChangeTagsForResource -UntagInTagsElem=RemoveTagKeys -UpdateTags -CreateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package route53
