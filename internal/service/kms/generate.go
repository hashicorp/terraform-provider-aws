//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListResourceTags -ListTagsInIDElem=KeyId -ServiceTagsSlice=yes -TagInIDElem=KeyId -TagTypeKeyElem=TagKey -TagTypeValElem=TagValue -UpdateTags=yes

package kms
