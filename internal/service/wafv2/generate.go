//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=ListIPSets,ListRegexPatternSets,ListRuleGroups,ListWebACLs -Paginator=NextMarker -Export=yes
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsInIDElem=ResourceARN -ListTagsOutTagsElem=TagInfoForResource.TagList -ServiceTagsSlice=yes -TagInIDElem=ResourceARN -UpdateTags=yes

package wafv2
