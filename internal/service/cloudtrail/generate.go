//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTags -ListTagsInIDElem=ResourceIdList -ListTagsInIDNeedSlice=yes -ListTagsOutTagsElem=ResourceTagList[0].TagsList -ServiceTagsSlice=yes -TagOp=AddTags -TagInIDElem=ResourceId -TagInTagsElem=TagsList -UntagOp=RemoveTags -UntagInNeedTagType=yes -UntagInTagsElem=TagsList -UpdateTags=yes

package cloudtrail
