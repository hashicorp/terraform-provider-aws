// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=ListActivatedRulesInRuleGroup,ListByteMatchSets,ListGeoMatchSets,ListIPSets,ListRateBasedRules,ListRegexMatchSets,ListRegexPatternSets,ListRules,ListRuleGroups,ListSizeConstraintSets,ListSqlInjectionMatchSets,ListSubscribedRuleGroups,ListWebACLs,ListXssMatchSets -Paginator=NextMarker
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsInIDElem=ResourceARN -ListTagsOutTagsElem=TagInfoForResource.TagList -ServiceTagsSlice -TagInIDElem=ResourceARN -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package wafregional
