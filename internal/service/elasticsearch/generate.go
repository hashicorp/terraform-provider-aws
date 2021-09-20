//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTags -ListTagsInIDElem=ARN -ListTagsOutTagsElem=TagList -ServiceTagsSlice=yes -TagOp=AddTags -TagInIDElem=ARN -TagInTagsElem=TagList -UntagOp=RemoveTags -UpdateTags=yes

package elasticsearch
