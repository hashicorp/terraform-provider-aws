// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListCachePolicies -InputPaginator=Marker -OutputPaginator=CachePolicyList.NextMarker -- list_cache_policies_pages_gen.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListContinuousDeploymentPolicies -InputPaginator=Marker -OutputPaginator=ContinuousDeploymentPolicyList.NextMarker -- list_continuous_deployment_policies_pages_gen.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListFieldLevelEncryptionConfigs -InputPaginator=Marker -OutputPaginator=FieldLevelEncryptionList.NextMarker -- list_field_level_encryption_configs_pages_gen.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListFieldLevelEncryptionProfiles -InputPaginator=Marker -OutputPaginator=FieldLevelEncryptionProfileList.NextMarker -- list_field_level_encryption_profiles_pages_gen.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListFunctions -InputPaginator=Marker -OutputPaginator=FunctionList.NextMarker -- list_functions_pages_gen.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListKeyGroups -InputPaginator=Marker -OutputPaginator=KeyGroupList.NextMarker -- list_key_groups_pages_gen.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListOriginAccessControls -InputPaginator=Marker -OutputPaginator=OriginAccessControlList.NextMarker -- list_origin_access_controls_pages_gen.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListOriginRequestPolicies -InputPaginator=Marker -OutputPaginator=OriginRequestPolicyList.NextMarker -- list_origin_request_policies_pages_gen.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListRealtimeLogConfigs -InputPaginator=Marker -OutputPaginator=RealtimeLogConfigs.NextMarker -- list_realtime_log_configs_gen.go
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListResponseHeadersPolicies -InputPaginator=Marker -OutputPaginator=ResponseHeadersPolicyList.NextMarker -- list_response_headers_policies_pages_gen.go
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsInIDElem=Resource -ListTagsOutTagsElem=Tags.Items -ServiceTagsSlice "-TagInCustomVal=&awstypes.Tags{Items: Tags(updatedTags)}" -TagInIDElem=Resource "-UntagInCustomVal=&awstypes.TagKeys{Items: removedTags.Keys()}" -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package cloudfront
