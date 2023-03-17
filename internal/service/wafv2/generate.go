//go:generate go run ../../generate/listpages/main.go -ListOps=ListIPSets,ListRegexPatternSets,ListRuleGroups,ListWebACLs -Paginator=NextMarker -ContextOnly
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=ResourceARN -ListTagsOutTagsElem=TagInfoForResource.TagList -ServiceTagsSlice -TagInIDElem=ResourceARN -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package wafv2
