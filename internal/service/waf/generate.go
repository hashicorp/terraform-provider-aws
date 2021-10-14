//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=ListByteMatchSets,ListGeoMatchSets,ListIPSets,ListRateBasedRules,ListRegexMatchSets,ListRegexPatternSets,ListRuleGroups,ListRules,ListSizeConstraintSets,ListSqlInjectionMatchSets,ListWebACLs,ListXssMatchSets -Paginator=NextMarker -Export=yes
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsInIDElem=ResourceARN -ListTagsOutTagsElem=TagInfoForResource.TagList -ServiceTagsSlice=yes -TagInIDElem=ResourceARN -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package waf
