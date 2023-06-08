//go:generate go run ../../generate/listpages/main.go -ListOps=ListByteMatchSets,ListGeoMatchSets,ListIPSets,ListRateBasedRules,ListRegexMatchSets,ListRegexPatternSets,ListRuleGroups,ListRules,ListSizeConstraintSets,ListSqlInjectionMatchSets,ListWebACLs,ListXssMatchSets -Paginator=NextMarker
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=ResourceARN -ListTagsOutTagsElem=TagInfoForResource.TagList -ServiceTagsSlice -TagInIDElem=ResourceARN -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package waf
