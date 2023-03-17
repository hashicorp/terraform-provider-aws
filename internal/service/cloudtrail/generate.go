//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTags -ListTagsInIDElem=ResourceIdList -ListTagsInIDNeedSlice=yes -ListTagsOutTagsElem=ResourceTagList[0].TagsList -ServiceTagsSlice -TagOp=AddTags -TagInIDElem=ResourceId -TagInTagsElem=TagsList -UntagOp=RemoveTags -UntagInNeedTagType -UntagInTagsElem=TagsList -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package cloudtrail
