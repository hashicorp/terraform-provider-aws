//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsInIDElem=ResourceId -ListTagsOutTagsElem=TagList -ServiceTagsSlice=yes -TagOp=AddTagsToResource -TagInIDElem=ResourceId -TagResTypeElem=ResourceType -UntagOp=RemoveTagsFromResource -UpdateTags=yes

package ssm
