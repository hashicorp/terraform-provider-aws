//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=ResourceId -ListTagsOutTagsElem=ResourceTagSet.Tags -ServiceTagsSlice -TagOp=ChangeTagsForResource -TagInIDElem=ResourceId -TagInTagsElem=AddTags -TagResTypeElem=ResourceType -UntagOp=ChangeTagsForResource -UntagInTagsElem=RemoveTagKeys -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package route53
