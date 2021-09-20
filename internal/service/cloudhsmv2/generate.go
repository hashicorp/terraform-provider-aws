//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTags -ListTagsInIDElem=ResourceId -ListTagsOutTagsElem=TagList -ServiceTagsSlice=yes -TagInIDElem=ResourceId -TagInTagsElem=TagList -UntagInTagsElem=TagKeyList -UpdateTags=yes

package cloudhsmv2
